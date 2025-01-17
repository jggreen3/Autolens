// import type { NextConfig } from "next";

const nextConfig = {
  rewrites: async () => {
    return [
      {
        source: "/api/:path*",
        destination:
          process.env.NODE_ENV === "development"
            ? "http://127.0.0.1:8000/api/:path*"
            : process.env.BACKEND_URL + "/:path*",
      },
    ];
  },
};
export default nextConfig;
