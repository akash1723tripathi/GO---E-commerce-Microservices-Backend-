import os
import random
import subprocess
from datetime import date, datetime, time, timedelta


MIN_COMMITS = 12
MAX_COMMITS = 20
DEFAULT_START = date(2026, 8, 26)  #year-month-date
DEFAULT_END = date(2026, 8, 28)   #year-month-date

def get_commit_count():
    while True:
        answer = input(f"How many commits ({MIN_COMMITS}-{MAX_COMMITS}, default 16): ").strip()
        if not answer:
            return 16
        try:
            count = int(answer)
        except ValueError:
            print("Please enter a whole number.")
            continue
        if MIN_COMMITS <= count <= MAX_COMMITS:
            return count
        print(f"Please choose a number from {MIN_COMMITS} to {MAX_COMMITS}.")


def get_repo_path():
    while True:
        answer = input("Repository path (default current directory): ").strip() or "."
        if not os.path.isdir(os.path.join(answer, ".git")):
            print("That directory is not a Git repository.")
            continue
        return answer


def get_filename():
    return input("File to update (default data.txt): ").strip() or "data.txt"


def random_datetime(start_day=DEFAULT_START, end_day=DEFAULT_END):
    """Return a random local timestamp, inclusive of both calendar dates."""
    day = start_day + timedelta(days=random.randint(0, (end_day - start_day).days))
    seconds = random.randint(0, 23 * 60 * 60 + 59 * 60 + 59)
    return datetime.combine(day, time.min) + timedelta(seconds=seconds)


def run_git(args, repo_path, env=None):
    result = subprocess.run(
        ["git", *args], cwd=repo_path, env=env, text=True, capture_output=True
    )
    if result.returncode:
        raise RuntimeError(result.stderr.strip() or f"git {' '.join(args)} failed")


def make_commit(commit_date, repo_path, filename, number):
    filepath = os.path.join(repo_path, filename)
    with open(filepath, "a", encoding="utf-8") as data_file:
        data_file.write(f"Pseudo commit {number:02d}: {commit_date.isoformat()}\n")

    # The requested data file may be covered by a user's global ignore rules.
    run_git(["add", "-f", "--", filename], repo_path)
    env = os.environ.copy()
    git_date = commit_date.strftime("%Y-%m-%dT%H:%M:%S%z")
    env["GIT_AUTHOR_DATE"] = git_date
    env["GIT_COMMITTER_DATE"] = git_date
    run_git(["commit", "-m", f"chore: update data ({number:02d})"], repo_path, env)


def main():
    print("Create dated commits in data.txt from 2026-08-26 through 2026-08-28.")
    count = get_commit_count()
    repo_path = get_repo_path()
    filename = get_filename()

    if os.path.isabs(filename):
        raise ValueError("Please enter a filename relative to the repository.")

    print(f"\nCreating {count} commits in {os.path.abspath(repo_path)}")
    print(f"Updating {filename} with dates from {DEFAULT_START} through {DEFAULT_END}.\n")

    for number in range(1, count + 1):
        commit_date = random_datetime()
        print(f"[{number}/{count}] {commit_date:%Y-%m-%d %H:%M:%S}")
        make_commit(commit_date, repo_path, filename, number)

    push = input("Push these commits to origin now? [Y/n]: ").strip().lower()
    if push in ("", "y", "yes"):
        run_git(["push", "origin", "HEAD"], repo_path)
        print("Commits pushed successfully.")
    else:
        print("Commits were created locally; nothing was pushed.")


if __name__ == "__main__":
    main()
