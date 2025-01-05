// import type { NextConfig } from "next";

const nextConfig= {
  rewrites: async () => {
    return {
      fallback: [
        {
          source: '/api/:path*',
          destination: process.env.NODE_ENV === 'development' 
            ? 'http://localhost:8080/:path*'
            : process.env.BACKEND_URL + '/:path*',
        },
      ],
    }
  },
};

export default nextConfig;
