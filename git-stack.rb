class GitStack < Formula
  desc "Git stack management tool with interactive TUI"
  homepage "https://github.com/sam-phinizy/git-stack"
  version "0.3.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-amd64"
      sha256 "4007e3b29079e9617213274665b56f95b962411d1e2b1fa6acc41302d0c78bbb"
    end
    if Hardware::CPU.arm?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-macos-arm64"
      sha256 "175a43e6a251efffdf96cce8ac2daabb961ea78363ab0695e1e0d9d56586c717"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/sam-phinizy/git-stack/releases/download/v#{version}/git-stack-linux"
      sha256 "b293a4183236141551cbd78e89700d19edf241e42cec6e8d24eaba436f611dd1"
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