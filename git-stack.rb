class GitStack < Formula
  desc "Git stack management tool with interactive TUI"
  homepage "https://github.com/sam-phinizy/git-stack"
  version "0.3.4"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-amd64"
      sha256 "a4946048414277ef0cfa6474d339dd09b1491506e966c1823ad430513b092918"
    end
    if Hardware::CPU.arm?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-arm64"
      sha256 "355d7dadea185c421cddf0f8839767992b219722a467915fc558c6b4765e8f25"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-linux"
      sha256 "13090d4c44fd40009dce25724ef15652fa7c51d8976df72b2c75d946d46a942d"
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