#!/usr/bin/env python3
"""Benchmark runner: nesh vs bash/dash/zsh across Phase 4 categories.

Build bin/nesh first:  go build -o bin/nesh ./cmd/nesh
Run from repo root:    python3 benchmarks/run.py [runs]
"""
import os
import shutil
import subprocess
import sys
import time

RUNS = int(sys.argv[1]) if len(sys.argv) > 1 else 30
NESH = "bin/nesh"

# (category, nesh script, baseline script) — baselines run under every shell.
CATEGORIES = [
    ("loop",     "benchmarks/nesh_scripts/loop.nsh",     "benchmarks/bash_baseline/loop.sh"),
    ("vars",     "benchmarks/nesh_scripts/vars.nsh",     "benchmarks/bash_baseline/vars.sh"),
    ("fn",       "benchmarks/nesh_scripts/fn.nsh",       "benchmarks/bash_baseline/fn.sh"),
    ("strings",  "benchmarks/nesh_scripts/strings.nsh",  "benchmarks/bash_baseline/strings.sh"),
    ("control",  "benchmarks/nesh_scripts/control.nsh",  "benchmarks/bash_baseline/control.sh"),
    ("pipeline", "benchmarks/nesh_scripts/pipeline.nsh", "benchmarks/bash_baseline/pipeline.sh"),
    ("redirect", "benchmarks/nesh_scripts/redirect.nsh", "benchmarks/bash_baseline/redirect.sh"),
]


def shells():
    out = []
    for s in ["bash", "dash", "zsh"]:
        path = shutil.which(s) or (f"/bin/{s}" if os.path.exists(f"/bin/{s}") else None)
        if path:
            out.append((s, path))
    return out


def timeit(cmd, runs):
    times = []
    for _ in range(runs):
        t0 = time.perf_counter()
        r = subprocess.run(cmd, shell=True, capture_output=True)
        t1 = time.perf_counter()
        if r.returncode != 0:
            sys.exit(f"FAILED ({cmd}): {r.stderr.decode()}")
        times.append((t1 - t0) * 1000)
    times.sort()
    return times[len(times) // 2]


def main():
    shell_list = shells()
    print(f"{RUNS} runs per cell, median ms\n")
    header = f"{'category':<10}{'nesh':>10}"
    for s, _ in shell_list:
        header += f"{s:>10}"
    print(header)
    print("-" * len(header))

    wins = 0
    for cat, nsh_script, base_script in CATEGORIES:
        nesh_ms = timeit(f"{NESH} {nsh_script}", RUNS)
        row = f"{cat:<10}{nesh_ms:>10.1f}"
        best = nesh_ms
        best_name = "nesh"
        results = {}
        for s, path in shell_list:
            ms = timeit(f"{path} {base_script}", RUNS)
            results[s] = ms
            if ms < best:
                best, best_name = ms, s
        for s, _ in shell_list:
            row += f"{results[s]:>10.1f}"
        ratio = max(nesh_ms, best) / min(nesh_ms, best)
        print(row + f"   {best_name} ({ratio:.1f}x)")
        if best_name == "nesh":
            wins += 1

    print(f"\nnesh fastest in {wins}/{len(CATEGORIES)} categories")
    subprocess.run("rm -f bench_out.tmp", shell=True)


if __name__ == "__main__":
    main()
