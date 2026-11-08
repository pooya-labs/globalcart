import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "GlobalCart — Project Case Study",
  description: "The GlobalCart 1.0.0 bilingual portfolio case study: marketplace UX, offline catalog, labelled demo pricing, engineering architecture and verified release quality.",
  alternates: { canonical: "https://pooya-labs.github.io/globalcart/case-study/" }
};

export default function CaseStudyLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return children;
}
