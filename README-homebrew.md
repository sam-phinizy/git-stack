# Homebrew Installation

This repo includes a Homebrew formula that downloads pre-built binaries from GitHub releases.

## Quick Install (Local)
```bash
brew install --formula ./git-stack.rb
```

## Setting up a Homebrew Tap
1. Create a new repository named `homebrew-tap`
2. Move `git-stack.rb` to `Formula/git-stack.rb` in that repo
3. Users can then install with:
```bash
brew tap sam-phinizy/tap
brew install git-stack
```

## Updating Checksums for New Releases

When you create a new release, you need to update the SHA256 checksums in the formula:

### 1. Create a tagged release
```bash
git tag v0.1.0
git push origin v0.1.0
```

### 2. Wait for the GitHub Actions workflow to complete

### 3. Generate SHA256 checksums
```bash
# For macOS Intel
curl -sL https://github.com/sam-phinizy/git-stack/releases/download/v0.1.0/git-stack-macos-amd64 | shasum -a 256

# For macOS Apple Silicon  
curl -sL https://github.com/sam-phinizy/git-stack/releases/download/v0.1.0/git-stack-macos-arm64 | shasum -a 256

# For Linux
curl -sL https://github.com/sam-phinizy/git-stack/releases/download/v0.1.0/git-stack-linux | shasum -a 256
```

### 4. Update the formula
Replace the placeholder SHA256 values in `git-stack.rb`:

```ruby
on_macos do
  if Hardware::CPU.intel?
    url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-amd64"
    sha256 "actual_intel_sha256_here"
  end
  if Hardware::CPU.arm?
    url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-arm64"
    sha256 "actual_arm_sha256_here"
  end
end

on_linux do
  if Hardware::CPU.intel?
    url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-linux"
    sha256 "actual_linux_sha256_here"
  end
end
```

### 5. Update the version number
Change the version in the formula to match your release tag (without the 'v' prefix).

## Example Complete Update

For version v0.1.1 with example checksums:

```ruby
class GitStack < Formula
  desc "Git stack management tool with interactive TUI"
  homepage "https://github.com/sam-phinizy/git-stack"
  version "0.1.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-amd64"
      sha256 "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
    end
    if Hardware::CPU.arm?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-arm64"
      sha256 "b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef1234567a"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-linux"
      sha256 "c3d4e5f6789012345678901234567890abcdef1234567890abcdef1234567ab2"
    end
  end

  def install
    binary_name = if OS.mac?
      if Hardware::CPU.intel?
        "git-stack-macos-amd64"
      else
        "git-stack-macos-arm64"
      end
    else
      "git-stack-linux"
    end
    bin.install binary_name => "git-stack"
  end

  test do
    system "#{bin}/git-stack", "--help"
  end
end
```