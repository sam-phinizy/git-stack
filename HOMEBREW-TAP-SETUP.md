# Setting Up Homebrew Tap: beer-hall

This guide will help you create a Homebrew tap called "beer-hall" for distributing your git-stack formula.

## Step 1: Create the Tap Repository

1. **Create a new repository on GitHub:**
   - Repository name: `beer-hall`
   - Owner: `sam-phinizy`
   - Make it public
   - Initialize with README

2. **Clone the repository locally:**
   ```bash
   git clone https://github.com/sam-phinizy/beer-hall.git
   cd beer-hall
   ```

## Step 2: Set Up Repository Structure

1. **Create the Formula directory:**
   ```bash
   mkdir Formula
   ```

2. **Copy the formula file:**
   ```bash
   # Copy the git-stack.rb from your main repository
   cp /path/to/git-stack/git-stack.rb Formula/git-stack.rb
   ```

3. **Create initial commit:**
   ```bash
   git add Formula/git-stack.rb
   git commit -m "Add git-stack formula"
   git push origin main
   ```

## Step 3: Update Main Repository to Push to Tap

Add this job to your `.github/workflows/release.yml` in the main git-stack repository, after the `update-homebrew-formula` job:

```yaml
  update-tap:
    needs: update-homebrew-formula
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/')
    
    steps:
      - name: Checkout main repo
        uses: actions/checkout@v4
        with:
          ref: main  # Get the updated formula from main branch

      - name: Checkout tap repository
        uses: actions/checkout@v4
        with:
          repository: sam-phinizy/beer-hall
          token: ${{ secrets.GITHUB_TOKEN }}
          path: beer-hall

      - name: Copy updated formula to tap
        run: |
          cp git-stack.rb beer-hall/Formula/git-stack.rb

      - name: Commit and push to tap
        run: |
          cd beer-hall
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add Formula/git-stack.rb
          git commit -m "Update git-stack formula to v${{ steps.version.outputs.version }}" || exit 0
          git push
```

## Step 4: User Installation Instructions

Once set up, users can install git-stack using your tap:

```bash
# Add the tap
brew tap sam-phinizy/beer-hall

# Install git-stack
brew install git-stack

# Or do it in one command
brew install sam-phinizy/beer-hall/git-stack
```

## Step 5: Repository Structure

Your final `beer-hall` repository should look like this:

```
beer-hall/
├── README.md
└── Formula/
    └── git-stack.rb
```

## Step 6: Update README.md in Tap Repository

Create a README.md in the beer-hall repository:

```markdown
# Sam Phinizy's Homebrew Tap (beer-hall)

This tap contains formulae for Sam Phinizy's tools.

## Installation

```bash
brew tap sam-phinizy/beer-hall
```

## Available Formulae

### git-stack
Git stack management tool with interactive TUI

```bash
brew install git-stack
```

## Usage

After installation, you can use:
- `git-stack` - Interactive git stack management

## Issues

Report issues at the main repository: https://github.com/sam-phinizy/git-stack/issues
```

## Verification

After setup, test that everything works:

1. **Test locally:**
   ```bash
   brew tap sam-phinizy/beer-hall
   brew install git-stack
   git-stack --help
   ```

2. **Test updates:**
   - Create a new tag in your main repository
   - Verify that both the main repo formula and tap formula get updated automatically

## Notes

- The tap name "beer-hall" will be referenced as `sam-phinizy/beer-hall`
- Users will install with `brew install sam-phinizy/beer-hall/git-stack` or just `brew install git-stack` after tapping
- The automated workflow ensures your tap stays in sync with releases
- Formula updates happen automatically on each tagged release

## Troubleshooting

- **Permission issues:** Make sure your GitHub token has write access to the beer-hall repository
- **Formula not found:** Ensure the file is in `Formula/git-stack.rb` (case-sensitive)
- **Installation fails:** Check that SHA256 checksums in the formula are correct