/**
 * Synthetic supplier payload for adapter contract tests only.
 * These records are not live inventory and must never be presented as such.
 */
export const syntheticSupplierRecords: unknown[] = [
  {
    productCode: 9001,
    displayName: "Demo travel bottle",
    image: "https://cdn.example.test/globalcart/demo-bottle.jpg",
    department: "accessories",
    unitPrice: 12.5,
    availableQuantity: 18,
    options: [
      {
        variantCode: "BOTTLE-500-BLUE",
        merchantSku: "GC-BTL-500-BLU",
        label: "500 ml / Blue",
        unitPrice: 13.25,
        availableQuantity: 7,
        specifications: { capacity: "500 ml", color: "blue" },
        enabled: true
      },
      {
        variantCode: "BOTTLE-750-BLACK",
        merchantSku: "GC-BTL-750-BLK",
        label: "750 ml / Black",
        unitPrice: 15.75,
        availableQuantity: 11,
        specifications: { capacity: "750 ml", color: "black" },
        enabled: false
      }
    ]
  }
];
