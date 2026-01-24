class ClaudeInsightsAgent < Formula
  desc "Sync Claude Code sessions to team server"
  homepage "https://github.com/dkd-dobberkau/claude-insights-agent"
  version "0.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/dkd-dobberkau/claude-insights-agent/releases/download/v0.2.0/claude-insights-agent-darwin-arm64"
      sha256 "1c813cc37f132b83adbaa9980f0ec246c891ebd930a47f1f232c76a063b70038"

      def install
        bin.install "claude-insights-agent-darwin-arm64" => "claude-insights-agent"
      end
    else
      url "https://github.com/dkd-dobberkau/claude-insights-agent/releases/download/v0.2.0/claude-insights-agent-darwin-amd64"
      sha256 "603e9cc828ced89e6a9566f17d2db47bf9091c9f159472b80bd1cc06a587c617"

      def install
        bin.install "claude-insights-agent-darwin-amd64" => "claude-insights-agent"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/dkd-dobberkau/claude-insights-agent/releases/download/v0.2.0/claude-insights-agent-linux-arm64"
      sha256 "ac6140b3dd2eb114bc53d7252e9e9e6142d75c3c22aca2ff19a4e17a8e5a5f7f"

      def install
        bin.install "claude-insights-agent-linux-arm64" => "claude-insights-agent"
      end
    else
      url "https://github.com/dkd-dobberkau/claude-insights-agent/releases/download/v0.2.0/claude-insights-agent-linux-amd64"
      sha256 "0382a8c0355552061184bfd8205c2de6019101b910f1c421c82852fdfd6b96dc"

      def install
        bin.install "claude-insights-agent-linux-amd64" => "claude-insights-agent"
      end
    end
  end

  def caveats
    <<~EOS
      To get started, run:
        claude-insights-agent init

      To run as a background service:
        brew services start claude-insights-agent
    EOS
  end

  service do
    run [opt_bin/"claude-insights-agent", "run"]
    keep_alive true
    log_path var/"log/claude-insights-agent.log"
    error_log_path var/"log/claude-insights-agent.log"
  end

  test do
    assert_match "claude-insights-agent v#{version}", shell_output("#{bin}/claude-insights-agent version")
  end
end
