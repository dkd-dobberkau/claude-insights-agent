class ClaudeInsightsAgent < Formula
  desc "Sync Claude Code sessions to team server"
  homepage "https://github.com/dkd-dobberkau/claude-insights-agent"
  version "0.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/dkd-dobberkau/claude-insights-agent/releases/download/v0.2.0/claude-insights-agent-darwin-arm64"
      sha256 "e15d335ac5dc401fc569b853797af52e30a65351f2e359c92227faf350142239"

      def install
        bin.install "claude-insights-agent-darwin-arm64" => "claude-insights-agent"
      end
    else
      url "https://github.com/dkd-dobberkau/claude-insights-agent/releases/download/v0.2.0/claude-insights-agent-darwin-amd64"
      sha256 "f3dbf3e16475abbcdb6b8f553ab180c64051092c92d2f070409ff1f77f4eb3c7"

      def install
        bin.install "claude-insights-agent-darwin-amd64" => "claude-insights-agent"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/dkd-dobberkau/claude-insights-agent/releases/download/v0.2.0/claude-insights-agent-linux-arm64"
      sha256 "6a930850d9583a687ffc0f7e35fd68e16d66291efd3d1adff17706195c930d3f"

      def install
        bin.install "claude-insights-agent-linux-arm64" => "claude-insights-agent"
      end
    else
      url "https://github.com/dkd-dobberkau/claude-insights-agent/releases/download/v0.2.0/claude-insights-agent-linux-amd64"
      sha256 "2bdb2641b82a1b5cfe10e83fea8c5e32c3d4063e2cbb9d51de2818d9938e4b6b"

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
