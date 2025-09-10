# git-stack

A utility to manage a stack of checked-out Git branches.

`git-stack` simplifies workflows with dependent branches by providing a set of commands to manage a stack of Git branches.

## Installation

### Homebrew

```sh
brew install sam-phinizy/beer-hall/git-stack
```

## Usage

`git-stack` provides the following commands:

| Command | Description |
| --- | --- |
| `git-stack` or `git-stack list` | Displays all branches currently in the stack. |
| `git-stack checkout <branch>` | Pushes the current branch onto the stack and checks out `<branch>`. |
| `git-stack pop` | Pops the last branch from the stack and checks it out. |
| `git-stack pick` | Interactively pick a branch from the stack to checkout. |
| `git-stack peek` | Shows the branch at the top of the stack. |
| `git-stack clear` | Clears all branches from the stack. |
| `git-stack up` | Checks out the next branch up in the stack. |
| `git-stack down` | Checks out the previous branch down in the stack. |
| `git-stack rebase` | Rebases the entire stack. |

### Global Flags

* `--stash`: Auto-stash local changes before switching branches (defaults to `true`).

### Rebase Command Flags

* `--pull`: Pull latest changes on the base branch before rebasing (defaults to `true`).
* `--continue`: Continue a stacked rebase after resolving conflicts.

## How it Works

`git-stack` works by storing the branch stack in a file named `git_branch_stack` inside your repository's `.git` directory.

## Building from Source

To build `git-stack` from source, you need to have Go installed.

1. Clone the repository:
   ```sh
   git clone https://github.com/sam-phinizy/git-stack.git
   ```
2. Navigate to the project directory:
   ```sh
   cd git-stack
   ```
3. Build the binary:
   ```sh
   go build -o git-stack .
   ```

## License

This project is licensed under the MIT License.
