require_relative '../pr-ready/comparison-verification/bench/comparison_support'
require_relative '../pr-ready/comparison-verification/bench/http_client'
require 'digest'
require 'time'
require 'net/http'
include BenchmarkSupport
root = Dir.pwd
seed = '/dev/shm/campfire-pr-ready/seed'
out = File.join(root, '.cache/post-fix-20261010/ordinal-cable')
raise 'output exists' if File.exist?(out)
FileUtils.mkdir_p(out)
labels = JSON.parse(File.read(File.join(seed, 'labels.json')))
fixture_env = File.readlines(File.join(root, '.cache/shared-current/reference.env'), chomp: true).reject { |l| l.empty? || l.start_with?('#') }.to_h { |l| l.split('=', 2) }
loadgen = File.join(root, '.cache/shared-verification-current/.cache/linux-target/release/loadgen')
room = Integer(labels.fetch('rooms.watercooler'))
people = 64
base = 'http://127.0.0.1:25160'
container = "cf-distinct-fanout-#{Process.pid}"
metadata = { started_at: Time.now.utc.iso8601, people: people, server_cpus: '0-3', client_cpus: '4-7', seed_sha256: Digest::SHA256.file(File.join(seed, 'db/production.sqlite3')).hexdigest, loadgen_sha256: Digest::SHA256.file(loadgen).hexdigest, source_revision: run('git', 'rev-parse', 'HEAD').strip, phases: {latency_msgs: 100, interval_ms: 100, tput_secs: 5, posters: 4, refresh_secs: 50}, images: {} }
results = []
sql = ->(db, query) { text = run('sqlite3', '-cmd', '.timeout 10000', '-json', db, query); text.strip.empty? ? [] : JSON.parse(text) }
lg = ->(*args) { JSON.parse(run('taskset', '-c', '4-7', loadgen, *args)) }
begin
  3.times do |round|
    order = round.even? ? %w[current ordinal] : %w[ordinal current]
    order.each do |app|
      data = "/dev/shm/campfire-pr-ready/distinct-fanout-runtime/#{Process.pid}/#{app}-#{round+1}"
      prepare_storage(seed, data)
      FileUtils.mkdir_p(File.join(data, 'logs'))
      db = File.join(data, 'db/production.sqlite3')
      david = labels.fetch('emails.david').gsub("'", "''")
      sql.call(db, "UPDATE push_subscriptions SET endpoint='https://127.0.0.1:9/push/'||id; UPDATE webhooks SET url='http://127.0.0.1:9/hook/'||id;")
      sql.call(db, "WITH RECURSIVE n(i) AS (VALUES(1) UNION ALL SELECT i+1 FROM n WHERE i<#{people}) INSERT INTO users(name,email_address,password_digest,role,status,created_at,updated_at) SELECT 'Fanout '||i,'fanout-'||i||'@example.test',password_digest,0,0,'2026-10-10 00:00:00','2026-10-10 00:00:00' FROM n CROSS JOIN users WHERE email_address='#{david}';")
      sql.call(db, "INSERT INTO memberships(room_id,user_id,involvement,created_at,updated_at) SELECT #{room},id,'everything','2026-10-10 00:00:00','2026-10-10 00:00:00' FROM users WHERE email_address LIKE 'fanout-%@example.test';")
      verified = sql.call(db, "SELECT COUNT(*) n,COUNT(DISTINCT u.id) distinct_users FROM users u JOIN memberships m ON m.user_id=u.id WHERE u.email_address LIKE 'fanout-%@example.test' AND u.role=0 AND u.status=0 AND m.room_id=#{room}").first
      raise 'not distinct active people' unless verified.values.all? { |n| n == people }
      initial = sql.call(db, 'SELECT MAX(id) id FROM messages').first.fetch('id')
      image = {'current'=>'campfire-post-fix','ordinal'=>'campfire-ordinal'}.fetch(app)
      metadata[:images][app] = run('docker','inspect','--format','{{.Id}}',image).strip
      config = fixture_env.merge('HTTP_PORT'=>'25160','TARGET_PORT'=>'25161','GOMAXPROCS'=>'4','TOKIO_WORKER_THREADS'=>'4','RAILS_MAX_THREADS'=>'4','JOB_CONCURRENCY'=>'3','GOGC'=>'100')
      run('chown','-R','1000:1000',data)
      run('docker','run','-d','--name',container,'--network','host','--cpuset-cpus','0-3',*environment(config),*mounts(File.join(data,'db')=>'/rails/storage/db', File.join(data,'files')=>'/rails/storage/files', File.join(data,'logs')=>'/rails/storage/logs'), image)
      deadline = clock + 60
      client = BenchmarkHTTPClient.new(base)
      until client.ready?
        raise 'server not ready' if clock > deadline
        sleep 0.1
      end
      run('docker','exec','--user','root',container,'chmod','-R','a+rwX','/rails/storage/db')
      cookie = lg.call('login','--base',base,'--email',labels.fetch('emails.david'),'--password',labels.fetch('passwords.all')).fetch('cookie')
      scrape = lg.call('scrape','--base',base,'--cookie',cookie,'--room',room.to_s)
      session_path = File.join(data, 'sessions.tsv')
      # Independent fixture people log in from independent loopback addresses;
      # leave the real per-IP authentication limiter enabled.
      login_person = ->(n) do
        http = Net::HTTP.new('127.0.0.1',25160,nil)
        http.local_host = "127.0.0.#{n+2}"
        jar = {}
        merge = ->(response) { (response.get_fields('set-cookie') || []).each { |s| k,v=s.split(';',2).first.split('=',2); jar[k]=v } }
        http.start do |h|
          page = h.get('/session/new'); merge.call(page)
          token = page.body[/<meta name="csrf-token" content="([^"]*)"/,1].to_s
          request = Net::HTTP::Post.new('/session')
          request['Cookie'] = jar.map { |k,v| "#{k}=#{v}" }.join('; ')
          request['Sec-Fetch-Site'] = 'same-origin'
          request.set_form_data({'email_address'=>"fanout-#{n+1}@example.test",'password'=>labels.fetch('passwords.all'),'authenticity_token'=>token})
          response = h.request(request); merge.call(response)
          raise "fixture login failed #{response.code}" unless response.code == '302' && jar.key?('session_token')
        end
        jar.map { |k,v| "#{k}=#{v}" }.join('; ')
      end
      File.open(session_path,'w',0600) do |file|
        people.times do |n|
          c = login_person.call(n)
          page = lg.call('scrape','--base',base,'--cookie',c,'--room',room.to_s)
          raise 'bad member room' unless page.fetch('status') == 200 && page.fetch('streams').length == 3
          identifiers = [{'channel'=>'PresenceChannel','room_id'=>room},{'channel'=>'UnreadRoomsChannel'},{'channel'=>'HeartbeatChannel'}].map { |s| JSON.generate(s) }
          page.fetch('streams').each { |s| channel,name=s.split('|',2); identifiers << JSON.generate({'channel'=>channel,'signed_stream_name'=>name}) }
          file.puts([c,*identifiers].join("\t"))
        end
      end
      authenticated = sql.call(db,"SELECT COUNT(DISTINCT s.user_id) n FROM sessions s JOIN users u ON u.id=s.user_id WHERE u.email_address LIKE 'fanout-%@example.test'").first.fetch('n')
      raise 'sessions do not own distinct users' unless authenticated == people
      pid = run('docker','inspect','--format','{{.State.Pid}}',container).strip
      memory = -> { File.read("/proc/#{pid}/status").lines.grep(/^(VmRSS|VmHWM|Threads):/).to_h { |l| k,v=l.split(':',2); [k,v.strip] } }
      samples = []
      argv = ['taskset','-c','4-7',loadgen,'cable','--base',base,'--cookie',cookie,'--room',room.to_s,'--csrf',scrape.fetch('csrf').to_s,'--clients',people.to_s,'--sessions',session_path,'--latency-msgs','100','--interval-ms','100','--tput-secs','5','--posters','4','--refresh-secs','50']
      started = clock
      cpu_before = Process.times
      stdout = nil
      Open3.popen3(*argv) do |stdin,output,error,wait|
        stdin.close
        reader = Thread.new { output.read }
        File.open(File.join(out,"#{app}-#{round+1}.phases.log"),'w') do |log|
          error.each_line do |line|
            log.write(line)
            samples << {phase: line.strip, memory: memory.call} if line.start_with?('PHASE')
          end
        end
        stdout = reader.value
        raise 'cable failed' unless wait.value.success?
      end
      sample = JSON.parse(stdout)
      run('ruby',File.join(root,'.cache/shared-verification-current/bench/check_sample.rb'),'cable',input:stdout)
      raise 'not all unique sessions used' unless sample.fetch('distinct_session_cookies') == people && sample.fetch('session_rows') == people
      sample['verified_distinct_people'] = authenticated
      sample['phase_memory'] = samples
      cpu_after = Process.times
      sample['generator_cpu_percent'] = 100*(cpu_after.cutime+cpu_after.cstime-cpu_before.cutime-cpu_before.cstime)/(clock-started)
      posted = sample.fetch('latency').fetch('post_successes') + sample.fetch('throughput').fetch('posted')
      actual = sql.call(db,"SELECT COUNT(*) n FROM messages WHERE id>#{initial}").first.fetch('n')
      bodies = sql.call(db,"SELECT rt.body FROM messages m JOIN action_text_rich_texts rt ON rt.record_type='Message' AND rt.record_id=m.id AND rt.name='body' WHERE m.id>#{initial} AND m.room_id=#{room}").map { |r| r.fetch('body') }
      seqs = bodies.map { |body| body[/fanout bmk([0-9]+)z/,1]&.to_i }.sort
      raise 'persisted marked writes differ' unless actual == posted && seqs == (1..posted).to_a
      indexed = sql.call(db,"SELECT COUNT(*) n FROM message_search_index WHERE rowid>#{initial} AND body MATCH 'fanout'").first.fetch('n')
      raise 'missing FTS writes' unless indexed == posted
      raise 'database integrity' unless sql.call(db,'PRAGMA integrity_check;').first.values == ['ok']
      deadline = clock + 10
      loop do
        connections = sql.call(db,"SELECT SUM(m.connections) n FROM memberships m JOIN users u ON u.id=m.user_id WHERE u.email_address LIKE 'fanout-%@example.test' AND m.room_id=#{room}").first.fetch('n')
        break if connections == 0
        raise 'presence not released after client shutdown' if clock > deadline
        sleep 0.1
      end
      sample['write_audit'] = {persisted: actual, exact_unique_markers: true, indexed: indexed, integrity: 'ok', presence_after_disconnect: 0}
      row = {app: app, round: round+1, cable: sample}
      results << row
      write_json(File.join(out,"#{app}-#{round+1}.json"),row)
      write_json(File.join(out,'results.json'),results)
      write_json(File.join(out,'metadata.json'),metadata)
      puts "#{app} #{round+1}: #{sample.dig('throughput','delivered_msgs_per_sec')} delivered/s; p99 paced all #{sample.dig('latency','all_clients','p99_ms')}ms; #{actual} audited writes"
      remove_container(container)
      FileUtils.rm_rf(data)
    end
  end
ensure
  remove_container(container)
end
