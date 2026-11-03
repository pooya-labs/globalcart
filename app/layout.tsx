import type { Metadata } from "next";
import "./globals.css";
export const metadata: Metadata = {
  title: "GlobalCart — Intelligent Marketplace",
  description: "A commerce-grade interactive marketplace portfolio by Pooya Goodarzi.",
  authors: [{ name: "Pooya Goodarzi", url: "https://github.com/pooya16pg-spec" }]
};
export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="fa" dir="rtl">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link rel="preconnect" href="https://images.unsplash.com" crossOrigin="anonymous" />
      </head>
      <body>{children}</body>
    </html>
  );
}
