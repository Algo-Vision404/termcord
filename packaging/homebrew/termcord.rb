class Termcord < Formula
  desc "Terminal-native Discord client — direct gateway, local cache"
  homepage "https://github.com/termcord/termcord"
  license "MIT"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/termcord/termcord/releases/download/v0.1.0/termcord_0.1.0_darwin_arm64.zip"
      sha256 "REPLACE_ON_RELEASE"
    else
      url "https://github.com/termcord/termcord/releases/download/v0.1.0/termcord_0.1.0_darwin_amd64.zip"
      sha256 "REPLACE_ON_RELEASE"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/termcord/termcord/releases/download/v0.1.0/termcord_0.1.0_linux_arm64.zip"
      sha256 "REPLACE_ON_RELEASE"
    else
      url "https://github.com/termcord/termcord/releases/download/v0.1.0/termcord_0.1.0_linux_amd64.zip"
      sha256 "REPLACE_ON_RELEASE"
    end
  end

  def install
    bin.install "termcord"
    bin.install "termcord-cli"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/termcord version")
  end
end
