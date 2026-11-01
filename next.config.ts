import type { NextConfig } from "next";
const onPages = process.env.GITHUB_ACTIONS === "true";
const config: NextConfig = {
  reactStrictMode: true,
  output: "export",
  trailingSlash: true,
  basePath: onPages ? "/globalcart" : "",
  images: { unoptimized: true }
};
export default config;
