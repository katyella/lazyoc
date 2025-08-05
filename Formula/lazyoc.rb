class Lazyoc < Formula
  desc "A lazy terminal UI for OpenShift/Kubernetes clusters"
  homepage "https://github.com/katyella/lazyoc"
  url "https://github.com/katyella/lazyoc/releases/download/v0.1.0/lazyoc_0.1.0_Darwin_#{Hardware::CPU.arch}.tar.gz"
  version "0.1.0"
  license "Apache-2.0"

  if Hardware::CPU.intel?
    sha256 "TO_BE_FILLED_BY_GORELEASER"
  elsif Hardware::CPU.arm?
    sha256 "TO_BE_FILLED_BY_GORELEASER"
  end

  depends_on "kubernetes-cli" => :optional
  
  def install
    bin.install "lazyoc"
  end

  test do
    system "#{bin}/lazyoc", "--version"
  end
end