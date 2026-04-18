/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [
      // Google Drive public image URLs
      {
        protocol: "https",
        hostname: "drive.google.com",
        pathname: "/uc",
      },
      // MinIO / S3-compatible storage
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
