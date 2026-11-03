import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "GlobalCart — Project Case Study",
  description: "A bilingual portfolio case study covering GlobalCart's product experience, data pipeline and backend architecture.",
  alternates: { canonical: "https://pooya-labs.github.io/globalcart/case-study/" }
};

export default function CaseStudyLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return children;
}
