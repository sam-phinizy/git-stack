class GitStack < Formula
  desc "Git stack management tool with interactive TUI"
  homepage "https://github.com/sam-phinizy/git-stack"
  version "0.3.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-amd64"
      sha256 "38e446191a4eb9e9327b575527e6b04bd421e3b4a8982538d49b30483865efa5"
    end
    if Hardware::CPU.arm?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-arm64"
      sha256 "e5b28b9d6418b33997e63bd7d776d95f8ee845045236546767702df515e2d8a8"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-linux"
      sha256 "409e1d035b7181ba5d164e6e153064c73f664d2959405a19bbc50edbe52da0ad"
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