# GlobalCart product catalog attribution

GlobalCart's local 1,000-product catalog is a selected and normalized subset of the **Amazon Berkeley Objects (ABO) Dataset**.

- Original dataset: https://amazon-berkeley-objects.s3.amazonaws.com/index.html
- Dataset registry: https://registry.opendata.aws/amazon-berkeley-objects/
- License: [Creative Commons Attribution 4.0 International (CC BY 4.0)](https://creativecommons.org/licenses/by/4.0/)
- Data and image credit: Amazon.com.
- Dataset authors: Matthieu Guillaumin, Thomas Dideriksen, Kenan Deng, Himanshu Arora, Arnab Dhua, Xi (Brian) Zhang, Tomas Yago-Vicente, Jasmine Collins, Shubham Goel, and Jitendra Malik.
- Metadata mirror used by the import pipeline: https://huggingface.co/datasets/hyper3labs/amazon-berkeley-objects

## Changes made for GlobalCart

The importer selects a balanced subset of 1,000 records from the original listings, maps product taxonomy to GlobalCart's category groups, and stores the selected small product images inside `public/images/catalog/`. Product records and images are committed and served as static site assets, so the storefront does not need a live third-party catalog request to render.

The source dataset does not provide retail prices, ratings, or stock figures in the listing fields used here. GlobalCart explicitly marks those fields as unavailable and does not fabricate them. The source product page is retained where an Amazon item identifier is present.

This reuse is attributed under CC BY 4.0. GlobalCart is an independent portfolio demonstration and is not endorsed by, affiliated with, or operated by Amazon.com or the University of California, Berkeley.
