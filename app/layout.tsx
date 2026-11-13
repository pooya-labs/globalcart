import type { Metadata } from "next";
import { DM_Sans, Vazirmatn } from "next/font/google";
import "./globals.css";
const dmSans = DM_Sans({ subsets: ["latin"], display: "swap", variable: "--font-dm-sans" });
const vazirmatn = Vazirmatn({ subsets: ["arabic", "latin"], display: "swap", variable: "--font-vazirmatn" });

export const metadata: Metadata = {
  title: "GlobalCart — Intelligent Marketplace",
  description: "An independent commerce experience owned by Pooya Labs and created by Pooya Goodarzi.",
  authors: [
    { name: "Pooya Labs", url: "https://github.com/pooya-labs" },
    { name: "Pooya Goodarzi", url: "https://github.com/pooya-labs" }
  ]
};
export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="fa" dir="rtl" className={`${dmSans.variable} ${vazirmatn.variable}`} suppressHydrationWarning>
      <head>
        {/* Route-level image preconnect hints live beside the routes that request them. */}
      </head>
      <body suppressHydrationWarning>{children}</body>
    </html>
  );
}
