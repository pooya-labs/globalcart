# GlobalCart product catalog attribution

GlobalCart's local 1,000-product catalog is a selected and normalized subset of the **Amazon Berkeley Objects (ABO) Dataset**.

- Original dataset: https://amazon-berkeley-objects.s3.amazonaws.com/index.html
- Dataset registry: https://registry.opendata.aws/amazon-berkeley-objects/
- License: [Creative Commons Attribution 4.0 International (CC BY 4.0)](https://creativecommons.org/licenses/by/4.0/)
- Data and image credit: Amazon.com.
- Dataset authors: Matthieu Guillaumin, Thomas Dideriksen, Kenan Deng, Himanshu Arora, Arnab Dhua, Xi (Brian) Zhang, Tomas Yago-Vicente, Jasmine Collins, Shubham Goel, and Jitendra Malik.
- Metadata mirror used by the import pipeline: https://huggingface.co/datasets/hyper3labs/amazon-berkeley-objects

## Changes made for GlobalCart

The importer selects a balanced subset of 1,000 records from the original listings and maps product taxonomy to GlobalCart's category groups. For each selected item it resolves the original `main_image_id` and all alternate `other_image_id` values against the official image metadata index, downloads the original photographs, optimizes them to WebP and stores the complete gallery in `public/images/catalog/`. The first image is the product-card/lead image; remaining images are kept in the product's `images[]` gallery. Product records and photos are committed and served as static site assets, so the storefront does not need a live third-party catalog or image request to render its data.

The source dataset does not provide retail prices, ratings, or stock figures in the listing fields used here. GlobalCart explicitly marks those fields as unavailable and does not fabricate them. The source item identifier is retained internally for data integrity and attribution; it is not rendered as an outbound product link. Product cards and detail pages are presented as GlobalCart catalog listings.

The original source images are optimized for web delivery; the image files in the repository are derivatives of the supplied original photographs. The storefront reserves image space, uses lazy loading and low fetch priority, and loads product photos independently from its page shell.\n\nThis reuse is attributed under CC BY 4.0. GlobalCart is an independent portfolio demonstration and is not endorsed by, affiliated with, or operated by Amazon.com or the University of California, Berkeley.
