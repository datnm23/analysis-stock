/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  env: {
    API_GATEWAY_URL: process.env.API_GATEWAY_URL || 'http://localhost:8080',
  },
};

module.exports = nextConfig;
