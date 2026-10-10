require_relative 'bench/comparison_support'
require_relative 'bench/http_client'
require_relative 'bench/contracts'
require 'digest'
require 'time'
include BenchmarkSupport
root = Dir.pwd
seed = File.join(root,'seed')
out = File.join(root,'native-fixed-post-v5')
raise 'output exists' if File.exist?(out)
FileUtils.mkdir_p(out)
labels = JSON.parse(File.read(File.join(seed,'labels.json')))
fixture_env = File.readlines(File.join(root,'reference.env'),chomp:true).reject { |l| l.empty? || l.start_with?('#') }.to_h { |l| l.split('=',2) }
loadgen = File.join(root,'loadgen-src/loadgen/target/release/loadgen')
rate_loadgen = File.join(root,'httprate')
room = Integer(labels.fetch('rooms.watercooler'))
write_room = Integer(labels.fetch('rooms.hq'))
base = 'http://127.0.0.1:25170'
container = "cf-fixed-latency-#{Process.pid}"
metadata = {started_at:Time.now.utc.iso8601,source_revision:'7aa66f2+uncommitted-client-watchdog-literal',seed_sha256:Digest::SHA256.file(File.join(seed,'db/production.sqlite3')).hexdigest,loadgen_sha256:Digest::SHA256.file(loadgen).hexdigest,rate_loadgen_sha256:Digest::SHA256.file(rate_loadgen).hexdigest,server_cpus:'0,2,4,6',client_cpus:'8,10,12,14',settings:{read_rate:10000,post_rate:3000,duration:8,warmup:2,concurrency_bound:64,gzip:1,gogc:100,pacing:'active'},validation:'Shared seed-grounded route contracts before timing and four shared per-response checks after each read sample. Timed httprate consumes full wire bodies and checks status/nonempty/transport, not semantic content. Each write sample independently audits exact persisted indexed body, unique client identity, room and creator; timed HTTP acknowledgment IDs are not captured.',images:{}}
results = []
sql = ->(db,query) { text=run('sqlite3','-cmd','.timeout 10000','-json',db,query);text.strip.empty? ? [] : JSON.parse(text) }
lg = ->(*args) { JSON.parse(run('taskset','-c','8,10,12,14',loadgen,*args)) }
begin
  3.times do |round|
    order = round.even? ? %w[go-before go] : %w[go go-before]
    order.each do |app|
      data = "/dev/shm/campfire-pr-ready/fixed-latency-runtime/#{Process.pid}/#{app}-#{round+1}"
      prepare_storage(seed,data)
      FileUtils.mkdir_p(File.join(data,'logs'))
      db = File.join(data,'db/production.sqlite3')
      sql.call(db,"UPDATE push_subscriptions SET endpoint='https://127.0.0.1:9/push/'||id; UPDATE webhooks SET url='http://127.0.0.1:9/hook/'||id;")
      image = {'go'=>'localhost/campfire-post-fix-desktop','go-before'=>'localhost/campfire-post-fix-desktop','rust'=>'campfire-current-rust'}.fetch(app)
      metadata[:images][app] = run('docker','inspect','--format','{{.Id}}',image).strip
      config = fixture_env.merge('HTTP_PORT'=>'25170','TARGET_PORT'=>'25171','GOMAXPROCS'=>'4','TOKIO_WORKER_THREADS'=>'4','RAILS_MAX_THREADS'=>'4','JOB_CONCURRENCY'=>'3','GOGC'=>'100')
      # Rootless container root maps to this benchmark user.
      run('docker','run','-d','--name',container,'--network','host','--security-opt','label=disable',*environment(config),*mounts(File.join(data,'db')=>'/rails/storage/db',File.join(data,'files')=>'/rails/storage/files',File.join(data,'logs')=>'/rails/storage/logs'),image,'taskset','-c','0,2,4,6',app == 'go' ? '/go-final' : '/go-upstream','server')
      deadline = clock + 60
      client = BenchmarkHTTPClient.new(base)
      until client.ready?
        if clock > deadline
          stdout,stderr,status=Open3.capture3('docker','logs',container)
          File.write(File.join(out,"#{app}-#{round+1}-startup.log"),stdout+stderr)
          raise 'server not ready'
        end
        sleep 0.1
      end
      run('docker','exec','--user','root',container,'chmod','-R','a+rwX','/rails/storage/db')
      sleep 3
      cookie = lg.call('login','--base',base,'--email',labels.fetch('emails.david'),'--password',labels.fetch('passwords.all')).fetch('cookie')
      scrape = lg.call('scrape','--base',base,'--cookie',cookie,'--room',room.to_s)
      prepared = BenchmarkContracts.prepare(base,cookie,db,labels,scrape.fetch('css'),File.join(data,'contracts'))
      routes = {'room_show'=>"/rooms/#{room}",'messages_page'=>"/rooms/#{room}/messages?before=#{labels.fetch('messages.busy_060')}",'sidebar'=>'/users/me/sidebar','search'=>'/searches?q=coffee','post_message'=>nil}
      samples = []
      routes.select { |name,path| name == 'post_message' }.each do |name,path|
        rate = path ? 10000 : 3000
        [2,8].each do |duration|
          before = sql.call(db,'SELECT MAX(id) id FROM messages').first.fetch('id')
          argv = ['taskset','-c','8,10,12,14',rate_loadgen,'--base',base,'--cookie',cookie,'--gzip','1','--conc','64','--pacing','active','--rate',rate.to_s,'--duration',duration.to_s]
          argv += path ? ['--path',path] : ['--post-room',write_room.to_s,'--csrf',scrape.fetch('csrf').to_s]
          started = clock
          cpu_before = Process.times
          sample = JSON.parse(run(*argv))
          cpu_after = Process.times
          sample['generator_cpu_percent'] = 100*(cpu_after.cutime+cpu_after.cstime-cpu_before.cutime-cpu_before.cstime)/(clock-started)
          sample['route'] = name
          sample['warmup'] = duration == 2
          write_json(File.join(out,"#{app}-#{round+1}-#{name}-#{duration}.json"),sample)
          raise 'lost offered arrivals' unless sample.fetch('errors') == 0 && sample.fetch('unscheduled') == 0 && sample.fetch('ok') == rate*duration && sample.fetch('statuses') == {'200'=>rate*duration}
          unless path
            rows = sql.call(db,"SELECT m.id,m.creator_id,m.room_id,m.client_message_id,rt.body,idx.body indexed_body FROM messages m JOIN action_text_rich_texts rt ON rt.record_type='Message' AND rt.record_id=m.id AND rt.name='body' JOIN message_search_index idx ON idx.rowid=m.id WHERE m.id>#{before}")
            creator = sql.call(db,"SELECT id FROM users WHERE email_address='#{labels.fetch('emails.david').gsub("'","''")}'").first.fetch('id')
            actual = sql.call(db,"SELECT COUNT(*) n FROM messages WHERE id>#{before}").first.fetch('n')
            bodies = rows.map { |r| r.fetch('body')[/bench write ([0-9]+)/,1]&.to_i }.sort
            raise 'persisted write population mismatch' unless actual == rate*duration && rows.length == actual && bodies == (0...actual).to_a
            raise 'wrong room, creator, identity or FTS' unless rows.map { |r| r.fetch('client_message_id') }.uniq.length == actual && rows.all? { |r| r.fetch('creator_id') == creator && r.fetch('room_id') == write_room && r.fetch('indexed_body').split.join(' ') == "bench write #{r.fetch('body')[/bench write ([0-9]+)/,1]}" }
            sample['persisted_write_audit'] = {count:actual,exact_bodies:true,unique_client_ids:true,correct_creator_room:true,normalized_fts:true}
          end
          samples << sample
          puts "#{app} #{round+1} #{name} #{duration}s: p99=#{sample.dig('latency','p99_ms')}ms service=#{sample.dig('service_latency','p99_ms')}ms late=#{sample.dig('generator_lateness','p99_ms')}ms"
        end
        if path
          probe = lg.call('http','--base',base,'--cookie',cookie,'--path',path,'--gzip','1','--conc','1','--duration','10','--requests','4','--validate',prepared.fetch(:contracts).fetch(name))
          run('ruby',File.join(root,'.cache/shared-verification-current/bench/check_sample.rb'),input:JSON.generate(probe))
          samples.last['after_shared_contract'] = probe
        end
      end
      raise 'database integrity' unless sql.call(db,'PRAGMA integrity_check;').first.values == ['ok']
      row = {app:app,round:round+1,http:samples,preflight:prepared.fetch(:preflight)}
      results << row
      write_json(File.join(out,"#{app}-#{round+1}.json"),row)
      write_json(File.join(out,'results.json'),results)
      write_json(File.join(out,'metadata.json'),metadata)
      remove_container(container)
      FileUtils.rm_rf(data)
    end
  end
ensure
  remove_container(container)
end
