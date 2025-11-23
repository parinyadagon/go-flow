import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        // ถ้าหน้าบ้านยิงมาที่ /api/...
        source: "/api/:path*",
        // ให้ส่งต่อไปหา Go Server ที่พอร์ต 8080
        destination: "http://localhost:8080/:path*",
      },
    ];
  },
};

export default nextConfig;
