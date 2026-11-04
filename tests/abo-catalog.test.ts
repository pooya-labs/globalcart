import test from "node:test";
import assert from "node:assert/strict";
import { ABO_SOURCE, classifyAboCategory, mapAboListing, stableAboProductId } from "../lib/catalog/abo";

const image = "https://amazon-berkeley-objects.s3.amazonaws.com/images/small/B000000001.jpg";

test("maps an ABO record with its original product identity and official image URL", () => {
  const product = mapAboListing({
    item_id: "B000000001",
    domain_name: "amazon.com",
    title: [{ language_tag: "de_DE", value: "Deutscher Titel" }, { language_tag: "en_US", value: "Stone & Beam Sofa" }],
    brand: [{ language_tag: "en_US", value: "Stone & Beam" }],
    product_type_readable: "Home Furniture",
    main_image_url: image
  }, "2026-11-04T00:00:00.000Z");

  assert.ok(product);
  assert.equal(product.source, ABO_SOURCE);
  assert.equal(product.sourceId, "B000000001");
  assert.equal(product.title, "Stone & Beam Sofa");
  assert.equal(product.brand, "Stone & Beam");
  assert.equal(product.category, "furniture");
  assert.equal(product.thumbnail, image);
  assert.equal(product.sourceUrl, "https://amazon.com/dp/B000000001");
  assert.equal(product.priceAvailable, false);
  assert.equal(product.ratingAvailable, false);
  assert.equal(product.stockAvailable, false);
  assert.equal(product.price, 0);
});

test("supports the official small-image path when a full URL is absent", () => {
  const product = mapAboListing({
    item_id: "B000000002",
    title: "Ceramic Kitchen Mug",
    product_type_readable: "Kitchen & Dining",
    main_image_path: "B000000002.jpg"
  });
  assert.ok(product);
  assert.equal(product.thumbnail, "https://amazon-berkeley-objects.s3.amazonaws.com/images/small/B000000002.jpg");
  assert.equal(product.category, "kitchen");
});

test("rejects missing identity, title, image and untrusted external image hosts", () => {
  assert.equal(mapAboListing({ item_id: "bad", title: "Product", main_image_url: image }), null);
  assert.equal(mapAboListing({ item_id: "B000000003", title: "Product" }), null);
  assert.equal(mapAboListing({ item_id: "B000000004", title: "Product", main_image_url: "https://evil.example/image.jpg" }), null);
});

test("classifies product families with accessory and category precedence", () => {
  assert.equal(classifyAboCategory({ title: "Phone Case for Galaxy S25", product_type: "Mobile Accessories" }), "mobile-accessories");
  assert.equal(classifyAboCategory({ title: "Galaxy S25 Smartphone", product_type: "Mobile Phones" }), "smartphones");
  assert.equal(classifyAboCategory({ title: "Wall Art Canvas", hierarchy_path: "Home > Wall Decor" }), "home-decor");
  assert.equal(classifyAboCategory({ title: "Sports Running Shoe", product_type_readable: "Footwear" }), "sports-outdoors");
  assert.equal(classifyAboCategory({ title: "Leather Tote", product_type_readable: "Handbags" }), "bags");
  assert.equal(classifyAboCategory({ title: "Unknown item" }), "other-products");
});

test("stable product identifiers are deterministic, positive and fit signed database integers", () => {
  const first = stableAboProductId("B000000009");
  assert.equal(first, stableAboProductId("B000000009"));
  assert.ok(first > 0 && first <= 2_147_483_647);
});


test("falls back to an official image identifier when the dataset omits the image path", () => {
  const product = mapAboListing({
    item_id: "B000000005",
    title: "Desk Lamp",
    product_type_readable: "Home Lighting",
    main_image_id: "B000000005"
  });
  assert.ok(product);
  assert.equal(product.thumbnail, "https://amazon-berkeley-objects.s3.amazonaws.com/images/small/B000000005.jpg");
});
