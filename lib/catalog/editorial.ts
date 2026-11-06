export type EditorialProductCopy = { title: string; description: string };

export const EDITORIAL_PRODUCT_GROUPS = {
  workspace: [601270869, 344744539, 2039335072, 2102718732],
  audio: [1721807718, 476309453, 400259401, 431773614],
  everyday: [2142599237, 467572772, 790152699, 870018017],
} as const;

const persianCopy: Record<number, EditorialProductCopy> = {
  601270869: { title: "پایه نگهدارنده لپ‌تاپ", description: "پایه‌ای جمع‌وجور برای بالاتر آوردن لپ‌تاپ و مرتب‌تر کردن فضای کار." },
  344744539: { title: "پایه رومیزی مانیتور", description: "پایه فلزی مانیتور برای ایجاد ارتفاع مناسب‌تر و فضای آزاد روی میز." },
  2039335072: { title: "میز قابل‌حمل لپ‌تاپ", description: "میز قابل‌تنظیم و جابه‌جاشدنی برای کار با لپ‌تاپ در فضاهای مختلف." },
  2102718732: { title: "هاب USB هفت‌درگاه", description: "هاب USB با هفت درگاه برای اتصال هم‌زمان لوازم جانبی سازگار." },
  1721807718: { title: "هدفون روگوشی سبک", description: "هدفون سیمی روگوشی با طراحی ساده و رنگ مشکی." },
  476309453: { title: "هدفون سیمی روگوشی", description: "هدفون روگوشی سیمی برای استفاده روزمره و گوش دادن به موسیقی." },
  400259401: { title: "شارژر دو درگاه همراه کابل", description: "شارژر دیواری دو درگاه USB به‌همراه کابل لایتنینگ؛ سازگاری را پیش از استفاده بررسی کن." },
  431773614: { title: "کابل لایتنینگ با روکش بافته‌شده", description: "کابل لایتنینگ به USB با روکش نایلونی بافته‌شده و گواهی MFi طبق عنوان منبع." },
  2142599237: { title: "کاور لپ‌تاپ ۱۳٫۳ اینچی", description: "کاور سبک برای محافظت و حمل لپ‌تاپ‌های سازگار تا اندازه ۱۳٫۳ اینچ." },
  467572772: { title: "کوله‌پشتی لپ‌تاپ ۱۵٫۶ اینچی", description: "کوله‌پشتی روزمره با محفظه مناسب لپ‌تاپ تا اندازه ۱۵٫۶ اینچ." },
  790152699: { title: "کوله‌پشتی کلاسیک مشکی", description: "کوله‌پشتی ساده و کاربردی برای استفاده روزمره و حمل وسایل شخصی." },
  870018017: { title: "ماوس بی‌سیم USB", description: "ماوس بی‌سیم با گیرنده نانو USB؛ رنگ قرمز." },
};

export function getPersianProductCopy(productId: number): EditorialProductCopy | undefined {
  return persianCopy[productId];
}

export function selectEditorialProducts<T extends { id: number }>(
  products: readonly T[],
  ids: readonly number[],
): T[] {
  const byId = new Map<number, T>(products.map(product => [product.id, product] as const));
  return ids.flatMap(id => {
    const product = byId.get(id);
    return product ? [product] : [];
  });
}
