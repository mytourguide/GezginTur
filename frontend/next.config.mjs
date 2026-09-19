/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [{ protocol: "https", hostname: "images.unsplash.com" }],
  },
  // Tarayici -> :8080 dogrudan istegini (CORS/pod sorunlari yasanabilir) atlatmak icin
  // arka uc cagirilari Next uzerinden proxy'lenir: /api/v1/* -> backend
  async rewrites() {
    return [
      { source: "/api/v1/:path*", destination: `${process.env.API_INTERNAL_URL || "http://localhost:8080"}/api/v1/:path*` },
    ];
  },
};

export default nextConfig;
