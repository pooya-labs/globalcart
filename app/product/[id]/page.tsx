import { notFound } from "next/navigation";
import catalogSnapshot from "../../../data/catalog.json";
import ProductDetailClient from "./ProductDetailClient";

type ProductVariant = {
  sku:string; title:string; price:number; stock:number;
  attributes?:Record<string,unknown>; source:string; sourceId:string; isActive:boolean;
};
type Product = {
  id:number; title:string; description:string; category:string; price:number;
  discountPercentage:number; rating:number; stock:number; brand?:string;
  thumbnail:string; images?:string[]; source?:string; sourceId?:string;
  importedAt?:string; variants?:ProductVariant[];
  priceAvailable?:boolean; priceKind?:"estimated"|"verified"; ratingAvailable?:boolean; stockAvailable?:boolean;
};

const products = catalogSnapshot.products as unknown as Product[];

export function generateStaticParams() {
  return products.map(product => ({ id: String(product.id) }));
}

export default async function ProductPage({ params }: { params: Promise<{ id:string }> }) {
  const { id } = await params;
  const product = products.find(item => String(item.id) === id);
  if (!product) notFound();
  const related = products
    .filter(item => item.category === product.category && item.id !== product.id)
    .slice(0,4);
  return <ProductDetailClient product={product} related={related} />;
}
