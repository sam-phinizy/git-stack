class GitStack < Formula
  desc "Git stack management tool with interactive TUI"
  homepage "https://github.com/sam-phinizy/git-stack"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-amd64"
      sha256 "SHA256_PLACEHOLDER_INTEL"
    end
    if Hardware::CPU.arm?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-arm64"
      sha256 "SHA256_PLACEHOLDER_ARM"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-linux"
      sha256 "SHA256_PLACEHOLDER"
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