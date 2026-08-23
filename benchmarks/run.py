#!/usr/bin/env python3
"""Simple A/B benchmark runner (no hyperfine needed)."""
import subprocess
import sys
import time

RUNS = 50


def bench(label, cmd):
    times = []
    for _ in range(RUNS):
        t0 = time.perf_counter()
        r = subprocess.run(cmd, shell=True, capture_output=True)
        t1 = time.perf_counter()
        if r.returncode != 0:
            sys.exit(f"{label} failed: {r.stderr.decode()}")
        times.append((t1 - t0) * 1000)
    times.sort()
    med = times[len(times) // 2]
    print(f"{label:12s} median {med:7.2f} ms   min {times[0]:7.2f} ms")


PAIRS = [
    ("nesh loop", "/tmp/opencode/nesh benchmarks/nesh_scripts/loop.nsh"),
    ("bash loop", "bash benchmarks/bash_baseline/loop.sh"),
    ("nesh vars", "/tmp/opencode/nesh benchmarks/nesh_scripts/vars.nsh"),
    ("bash vars", "bash benchmarks/bash_baseline/vars.sh"),
    ("nesh fn", "/tmp/opencode/nesh benchmarks/nesh_scripts/fn.nsh"),
    ("bash fn", "bash benchmarks/bash_baseline/fn.sh"),
]

for label, cmd in PAIRS:
    bench(label, cmd)
