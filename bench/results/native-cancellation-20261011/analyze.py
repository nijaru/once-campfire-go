import json
from pathlib import Path

base = Path(__file__).parent
for cohort in ["native-cable-boundary", "vm-boundary-capture"]:
    root = base / cohort
    metadata = json.loads((root / "metadata.json").read_text())
    ticks = (metadata.get("observation") or metadata["sampler"])["clock_ticks_per_second"]
    analyses = []
    for result in json.loads((root / "results.json").read_text()):
        stem = f"{result['app']}-{result['round']}"
        rows = json.loads((root / f"{stem}.proc.json").read_text())
        events = {}
        for line in (root / f"{stem}.phases.log").read_text().splitlines():
            if line.startswith("PHASE "):
                _, phase, timestamp = line.split()
                events[phase] = int(timestamp) * 1_000_000
        a = next(row for row in rows if row["phase"] == "saturated")
        b = next(row for row in rows if row["phase"] == "saturated_posted")
        wall = (b["monotonic_ns"] - a["monotonic_ns"]) / 1e9
        output = {
            "app": result["app"], "round": result["round"],
            "delivered_per_second": result["cable"]["throughput"]["delivered_msgs_per_sec"],
            "posted": result["cable"]["throughput"]["posted"],
            "wall_seconds": wall,
            "boundary_error_ms": [(row["unix_ns"] - events[row["phase"]]) / 1e6 for row in [a,b]],
            "observation_ms": [row["observation_ns"] / 1e6 for row in [a,b]],
        }
        for owner in ["server", "client"]:
            old, new = a[owner], b[owner]
            if old is None or new is None:
                output[owner] = {"available": False, "reason": "process snapshot unavailable (ENOENT/ESRCH)"}
                continue
            common = old["threads"].keys() & new["threads"].keys()
            cpu = (new["cpu_ticks"] - old["cpu_ticks"]) / ticks
            stats = {
                "available": True, "cpu_seconds": cpu, "cpu_cores": cpu / wall,
                "cpu_seconds_per_posted_message": cpu / output["posted"],
                "common_threads": len(common),
                "new_threads": len(new["threads"].keys() - old["threads"].keys()),
                "missing_threads": len(old["threads"].keys() - new["threads"].keys()),
                "runnable_wait_seconds_common_threads": 0.0,
                "scheduler_stat_threads": 0,
                "context_switches": {"voluntary": 0, "involuntary": 0},
                "cpu_max": old["cpu_max"],
            }
            for tid in common:
                ot, nt = old["threads"][tid], new["threads"][tid]
                if ot["runnable_wait_ns"] is not None and nt["runnable_wait_ns"] is not None:
                    stats["runnable_wait_seconds_common_threads"] += (nt["runnable_wait_ns"] - ot["runnable_wait_ns"]) / 1e9
                    stats["scheduler_stat_threads"] += 1
                for name, key in [("voluntary", "voluntary_ctxt_switches"), ("involuntary", "nonvoluntary_ctxt_switches")]:
                    stats["context_switches"][name] += nt["switches"][key] - ot["switches"][key]
            if old["cgroup_cpu"] is not None and new["cgroup_cpu"] is not None:
                stats["cgroup_cpu_delta"] = {key: new["cgroup_cpu"][key] - value for key,value in old["cgroup_cpu"].items()}
            output[owner] = stats
        for name, cpus in [("server_cpus", metadata["server_cpus"]), ("client_cpus", metadata["client_cpus"])]:
            # Cohorts use either a comma-separated CPU list or one contiguous range.
            numbers = list(map(int,cpus.split(","))) if "," in cpus else list(range(int(cpus.split("-")[0]),int(cpus.split("-")[-1])+1))
            total = busy = stolen = 0
            for cpu in numbers:
                delta = [bv-av for av,bv in zip(a["cpu"][f"cpu{cpu}"],b["cpu"][f"cpu{cpu}"])]
                total += sum(delta[:8])  # guest fields duplicate user/nice.
                busy += sum(delta[:3]) + sum(delta[5:8])
                stolen += delta[7]
            output[name] = {"busy_fraction": busy / total, "steal_fraction": stolen / total}
        analyses.append(output)
    (root / "phase-analysis.json").write_text(json.dumps(analyses,indent=2)+"\n")
    for row in analyses:
        print(cohort,row["app"],row["round"],"rate",row["delivered_per_second"],"cores",*[round(row[k]["cpu_cores"],3) if row[k]["available"] else None for k in ["server","client"]])
