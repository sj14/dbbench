class BrewFormula < Formula
  desc "dbbench is a simple database benchmarking tool which supports several databases"
  homepage ""
  url "https://github.com/sj14/dbbench/releases/download/v0.1.1/dbbench_0.1.1_darwin_amd64.tar.gz"
  version "0.1.1"
  sha256 "c4c88e4b65a1729ee9bff9bdc7fa52828005182a4eafa87b7ab88b5a76828073"

  def install
    bin.install "dbbench"
  end
end
