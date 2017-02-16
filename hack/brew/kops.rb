class Kops < Formula
  desc "Production Grade K8s Installation, Upgrades, and Management"
  homepage "https://github.com/kubernetes/kops"
  url "https://github.com/kubernetes/kops/archive/v1.4.1.tar.gz"
  sha256 "69b3c9d7e214109cfd197031091ed23963383c894e92804306629f6a32ab324b"
  head "https://github.com/kubernetes/kops.git"

  bottle do
    cellar :any_skip_relocation
    sha256 "99fc900bb11b242b4d3eca456fc7956233a8efa92f3dee7b321a005de0e94a28" => :sierra
    sha256 "ed71d71b5031e0918478dec06b9064cf7e3f5b907128e98ec85a187801a27f8e" => :el_capitan
    sha256 "0eee45caca5eb2a67ab88a90f9da226e99538d5198a196067b2a224a816bd6e0" => :yosemite
  end

  depends_on "go" => :build
  depends_on "kubernetes-cli"
  depends_on "md5sha1sum"

  def install
    ENV["VERSION"] = version unless build.head?
    ENV["GOPATH"] = buildpath
    kopspath = buildpath/"src/k8s.io/kops"
    kopspath.install Dir["*"]
    system "make", "-C", kopspath
    bin.install("bin/kops")
  end

  test do
    system "#{bin}/kops", "version"
  end
end
