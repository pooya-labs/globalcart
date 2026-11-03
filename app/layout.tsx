import type { Metadata } from "next";
import { DM_Sans, Vazirmatn } from "next/font/google";
import "./globals.css";
const dmSans = DM_Sans({ subsets: ["latin"], display: "swap", variable: "--font-dm-sans" });
const vazirmatn = Vazirmatn({ subsets: ["arabic", "latin"], display: "swap", variable: "--font-vazirmatn" });

export const metadata: Metadata = {
  title: "GlobalCart — Intelligent Marketplace",
  description: "A commerce-grade interactive marketplace portfolio by Pooya Goodarzi.",
  authors: [{ name: "Pooya Goodarzi", url: "https://github.com/pooya16pg-spec" }]
};
export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="fa" dir="rtl" className={`${dmSans.variable} ${vazirmatn.variable}`}>
      <head>
        {/* Route-level image preconnect hints live beside the routes that request them. */}
      </head>
      <body>{children}</body>
    </html>
  );
}
