import type { Metadata } from "next";
import "./globals.css";
export const metadata: Metadata = {
  title: "GlobalCart — Intelligent Marketplace",
  description: "A commerce-grade interactive marketplace portfolio by Pooya Goodarzi.",
  authors: [{ name: "Pooya Goodarzi", url: "https://github.com/pooya16pg-spec" }]
};
export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return <html lang="fa" dir="rtl"><body>{children}</body></html>;
}
