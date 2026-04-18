/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [
      // MinIO / S3-compatible storage — wildcard covers any configured endpoint
      {
        protocol: "http",
        hostname: "**",
        port: "**",
        pathname: "/blog-images/**",
      },
      {
        protocol: "https",
        hostname: "**",
        pathname: "/blog-images/**",
      },
    ],
  },
};

export default nextConfig;
