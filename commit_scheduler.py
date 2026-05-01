#!/usr/bin/env python3
"""
Commit Scheduler — Incremental commit push with configurable interval.

Pushes commits defined in commits.json to the remote (e.g. origin/main),
respecting a 2h + jitter (10–30 min) interval in scheduled mode.
Runs only during business hours: 08:00 to 19:00, Monday to Saturday (local time).
Uses local time. Does not add co-author. Can be run manually or in loop (--schedule).

Usage:
  python commit_scheduler.py              # push next commit now (manual)
  python commit_scheduler.py --manual     # same
  python commit_scheduler.py --schedule  # loop: commit -> wait 2h+jitter -> repeat
  python commit_scheduler.py --status    # show last pushed commit and next
"""

from __future__ import annotations

import argparse
import json
import os
import random
import subprocess
import sys
from datetime import datetime, timedelta
from pathlib import Path

# --- Configuration ---
_SCRIPT_DIR = Path(__file__).resolve().parent
# Repo root: este script na raiz do projeto, ou em scripts/ (commits.json na raiz)
REPO_ROOT = _SCRIPT_DIR if (_SCRIPT_DIR / "commits.json").exists() else _SCRIPT_DIR.parent
COMMITS_FILE = REPO_ROOT / "commits.json"
STATE_FILE = REPO_ROOT / ".strategy_state.json"
INTERVAL_BASE_MINUTES = 1.5 * 60  # 1.5 hours
JITTER_MIN_MINUTES = 10
JITTER_MAX_MINUTES = 30

# Business hours: only push during this window (local time)
BUSINESS_HOUR_START = 8   # 08:00
BUSINESS_HOUR_END = 19    # until 19:00 (exclusive: 18:59 is last minute)
BUSINESS_DAYS = (0, 1, 2, 3, 4, 5)  # Monday=0 to Saturday=5 (Sunday=6 excluded)
CHECK_BUSINESS_INTERVAL_WEEKEND_SECONDS = 60 * 60  # outside business hours: recheck every 60 min


def load_commits() -> list[dict]:
    """Load the commit list from commits.json."""
    if not COMMITS_FILE.exists():
        print(f"Error: {COMMITS_FILE} not found.", file=sys.stderr)
        sys.exit(1)
    with open(COMMITS_FILE, "r", encoding="utf-8") as f:
        data = json.load(f)
    if not isinstance(data, list):
        print("Error: commits.json must be a list of objects.", file=sys.stderr)
        sys.exit(1)
    return data


def load_state() -> dict:
    """Load the state of the last pushed commit."""
    if not STATE_FILE.exists():
        return {"last_pushed_index": -1, "last_push_timestamp": None, "last_pushed_message": None}
    try:
        with open(STATE_FILE, "r", encoding="utf-8") as f:
            return json.load(f)
    except (json.JSONDecodeError, OSError):
        return {"last_pushed_index": -1, "last_push_timestamp": None, "last_pushed_message": None}


def save_state(index: int, message: str) -> None:
    """Persist the index and message of the last pushed commit (local time)."""
    now = datetime.now().isoformat()
    state = {
        "last_pushed_index": index,
        "last_push_timestamp": now,
        "last_pushed_message": message,
    }
    with open(STATE_FILE, "w", encoding="utf-8") as f:
        json.dump(state, f, indent=2, ensure_ascii=False)


def run_git(args: list[str], cwd: Path | None = None) -> subprocess.CompletedProcess:
    """Run a git command in the repository."""
    cmd = ["git"] + args
    return subprocess.run(
        cmd,
        cwd=cwd or REPO_ROOT,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )


def is_tracked(path: str) -> bool:
    """True when path is tracked by git (even if deleted in working tree)."""
    r = run_git(["ls-files", "--error-unmatch", "--", path], cwd=REPO_ROOT)
    return r.returncode == 0


def get_next_commit_index(commits: list[dict], state: dict) -> int | None:
    """Return the index of the next commit to push, or None if none left."""
    last = state.get("last_pushed_index", -1)
    next_index = last + 1
    if next_index >= len(commits):
        return None
    return next_index


def push_one_commit(commits: list[dict], index: int) -> tuple[bool, bool]:
    """
    Add the commit files, commit (no co-author), and push.
    Returns (success, skipped): skipped=True when there was nothing to commit (already committed).
    """
    spec = commits[index]
    message = spec.get("message", "").strip()
    files = spec.get("files") or []

    if not message:
        print("Error: empty commit message.", file=sys.stderr)
        return (False, False)

    # Resolve paths relative to repository.
    # Include files that exist OR are tracked (so deletions can be committed).
    paths = []
    for p in files:
        rel_path = str(Path(p).as_posix())
        if rel_path.startswith("./"):
            rel_path = rel_path[2:]
        abs_path = (REPO_ROOT / rel_path).resolve()
        try:
            abs_path.relative_to(REPO_ROOT)
        except ValueError:
            continue
        if abs_path.exists() or is_tracked(rel_path):
            paths.append(rel_path)
        else:
            print(f"Warning: file not found/untracked (ignored): {rel_path}", file=sys.stderr)

    if not paths:
        print("Error: no valid files for this commit.", file=sys.stderr)
        return (False, False)

    # Stage only listed paths (including deletions with -A pathspec).
    add_result = run_git(["add", "-A", "--"] + paths, cwd=REPO_ROOT)
    if add_result.returncode != 0:
        print(f"Error adding files: {add_result.stderr}", file=sys.stderr)
        return (False, False)

    # Commit strictly only listed paths, even if other files are staged.
    commit_result = run_git(["commit", "-m", message, "--"] + paths, cwd=REPO_ROOT)
    if commit_result.returncode != 0:
        out_and_err = (commit_result.stdout + " " + commit_result.stderr).lower()
        if (
            "nothing to commit" in out_and_err
            or "working tree clean" in out_and_err
            or "no changes added to commit" in out_and_err
        ):
            # Already committed: advance state so re-run does not get stuck
            print("Nothing to commit (files already committed). Advancing state.", file=sys.stderr)
            return (True, True)
        print(f"Error committing: {commit_result.stderr or commit_result.stdout}", file=sys.stderr)
        return (False, False)

    # git push (current branch; default remote)
    push_result = run_git(["push"], cwd=REPO_ROOT)
    if push_result.returncode != 0:
        print(f"Error pushing: {push_result.stderr or push_result.stdout}", file=sys.stderr)
        return (False, False)

    return (True, False)


def is_business_hours(dt: datetime | None = None) -> bool:
    """True if within business hours: 08:00–19:00, Monday–Saturday (local time)."""
    t = dt or datetime.now()
    if t.weekday() not in BUSINESS_DAYS:
        return False
    return BUSINESS_HOUR_START <= t.hour < BUSINESS_HOUR_END


def wait_until_business_hours() -> None:
    """Wait in a loop until business hours (recheck every CHECK_BUSINESS_INTERVAL_WEEKEND_SECONDS)."""
    import time
    interval = CHECK_BUSINESS_INTERVAL_WEEKEND_SECONDS
    while not is_business_hours():
        now = datetime.now()
        print(f"[{now.strftime('%Y-%m-%d %H:%M:%S')}] Outside business hours (08:00–19:00, Mon–Sat). Next check in {interval // 60} min...")
        time.sleep(interval)


def next_interval_seconds() -> int:
    """Interval until next push: 2h + random between 10 and 30 minutes (in seconds)."""
    jitter = random.randint(JITTER_MIN_MINUTES, JITTER_MAX_MINUTES)
    return (INTERVAL_BASE_MINUTES + jitter) * 60


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Push commits incrementally (manual or scheduled). Local time, no co-author."
    )
    parser.add_argument(
        "--manual",
        action="store_true",
        help="Push only the next commit and exit (default).",
    )
    parser.add_argument(
        "--schedule",
        action="store_true",
        help="Scheduled mode: push one commit, wait 2h+jitter, repeat until done.",
    )
    parser.add_argument(
        "--status",
        action="store_true",
        help="Show last pushed commit and next in list.",
    )
    args = parser.parse_args()

    if args.status:
        commits = load_commits()
        state = load_state()
        last = state.get("last_pushed_index", -1)
        next_idx = get_next_commit_index(commits, state)
        print(f"Total commits in list: {len(commits)}")
        print(f"Last pushed (index): {last}")
        if last >= 0 and last < len(commits):
            print(f"Last message: {commits[last].get('message', '')}")
        if next_idx is not None:
            print(f"Next (index): {next_idx}")
            print(f"Next message: {commits[next_idx].get('message', '')}")
        else:
            print("No next commit (all already pushed).")
        return

    use_schedule = args.schedule
    commits = load_commits()
    state = load_state()

    while True:
        next_index = get_next_commit_index(commits, state)
        if next_index is None:
            print("All commits have been pushed.")
            break

        # In scheduled mode, push only during business hours (08:00–19:00, Mon–Sat)
        if use_schedule:
            wait_until_business_hours()

        spec = commits[next_index]
        message = spec.get("message", "")
        print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] Pushing commit {next_index + 1}/{len(commits)}: {message[:60]}...")

        success, skipped = push_one_commit(commits, next_index)
        if success:
            save_state(next_index, message)
            state = load_state()  # reload for next iteration (--schedule) to use correct index
            if skipped:
                print(f"Commit already present; state updated. Next: {next_index + 2}/{len(commits)}")
            else:
                print(f"Commit pushed: {message}")
        else:
            print("Push failed. Exiting.", file=sys.stderr)
            sys.exit(1)

        if not use_schedule:
            break

        delay = next_interval_seconds()
        next_run = datetime.now() + timedelta(seconds=delay)
        print(f"Next push in {delay // 60} min (around {next_run.strftime('%H:%M')}). Waiting...")
        try:
            import time
            time.sleep(delay)
        except KeyboardInterrupt:
            print("\nInterrupted by user.")
            break

    print("Done.")


if __name__ == "__main__":
    main()
