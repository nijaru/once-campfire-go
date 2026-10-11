# Private external observation only: no application or generator hooks.
module ProcSamples
  module_function

  def fields(path)
    File.read(path).rpartition(') ').last.split
  end

  def process(pid)
    path = "/proc/#{pid}"
    stat = fields("#{path}/stat")
    threads = Dir.glob("#{path}/task/*").filter_map do |task|
      begin
        state = fields("#{task}/stat")
        sched = File.file?("#{task}/schedstat") ? File.read("#{task}/schedstat").split.map(&:to_i) : nil
        switches = File.readlines("#{task}/status").grep(/^(voluntary_ctxt_switches|nonvoluntary_ctxt_switches):/).to_h { |line| key,value=line.split(':',2); [key,value.strip.to_i] }
        [File.basename(task), {state: state[0], cpu: state[36].to_i, runtime_ns: sched&.[](0), runnable_wait_ns: sched&.[](1), timeslices: sched&.[](2), switches: switches, wchan: File.read("#{task}/wchan").strip}]
      rescue Errno::ENOENT, Errno::ESRCH
        nil # Report missing threads in the analysis, not negative deltas.
      end
    end.to_h
    group = File.readlines("#{path}/cgroup").find { |line| line.start_with?('0::') }&.split('::',2)&.last&.strip
    cgroup = "/sys/fs/cgroup#{group}"
    stats = File.file?("#{cgroup}/cpu.stat") ? File.readlines("#{cgroup}/cpu.stat").to_h { |line| key,value = line.split; [key,value.to_i] } : nil
    {cpu_ticks: stat[11].to_i + stat[12].to_i, threads: threads, cgroup: group, cpu_max: File.file?("#{cgroup}/cpu.max") ? File.read("#{cgroup}/cpu.max").strip : nil, cgroup_cpu: stats,
     memory: File.readlines("#{path}/status").grep(/^(VmRSS|VmHWM|Threads):/).to_h { |line| key,value=line.split(':',2); [key,value.strip] }}
  rescue Errno::ENOENT, Errno::ESRCH
    nil
  end

  def snapshot(server, client, phase)
    start = Process.clock_gettime(Process::CLOCK_MONOTONIC, :nanosecond)
    row = {monotonic_ns: start, unix_ns: Process.clock_gettime(Process::CLOCK_REALTIME, :nanosecond), phase: phase,
           server: process(server), client: process(client),
           cpu: File.readlines('/proc/stat').grep(/^cpu\d+ /).to_h { |line| name,*values=line.split; [name,values.map(&:to_i)] }}
    row[:observation_ns] = Process.clock_gettime(Process::CLOCK_MONOTONIC, :nanosecond) - start
    row
  end
end
