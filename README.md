# Autolens - Computer Vision for Car Parts

Welcome to Autolens! This project is a computer vision application designed to detect and identify car parts in an image. It features a Next.js front end for a smooth user experience and a Go backend leveraging a fine-tuned YOLOv8 model for powerful, real-time object detection.

You can visit the live app on Vercel ([here](https://autolens.vercel.app/)).

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#architecture">Architecture</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->

## About the Project

As a DIY car enthusiast, I've often found myself stuck trying to identify a random part that I've removed—or worse, broken—while working on my car. Figuring out what a part is called or where to find a replacement can be a huge hassle. Autolens aims to solve this problem by applying computer vision to quickly and accurately identify car parts from any uploaded image.

### Architecture:

#### Object Detection and Classification:

The core of Autolens is a fine-tuned YOLOv8 model trained on a labeled dataset of around 10,000 images, spanning 50 unique car-part classes. For model development and fine-tuning, I used Python, Ultralytics, and PyTorch, which provided a robust ecosystem for training.

To deliver faster inference in production, the model was converted to ONNX format. A custom Go inference server handles image preprocessing, inference, and post-processing, providing high performance and scalability. The front end, built with Next.js, sends images to the Go backend for real-time predictions.

## Getting Started:

### Installation:

If you'd like to run this app locally, follow these steps to get it up and running:

1.  Clone the repository:

    ```shell
    git clone https://github.com/jggreen3/Autolens.git
    cd Autolens
    ```

2.  Install dependencies:

    ```shell
    yarn install
    ```

3.  Run the development server:

    ```shell
    yarn dev
    ```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## Roadmap

This is work in progress and there's a lot I'd like to do.

- <strong>Expand the Training Set:</strong> Add more images, including parts installed on cars, and diversify
  categories for broader coverage.
- <strong>Improve Deployment Infrastructure:</strong> Move away from serverless-only Vercel deployment
  for the backend. This would allow a more flexible setup without requiring a single monolithic file
  and reduce cold-start latency by avoiding frequent model downloads from S3.
- <strong>Service Manual Integration:</strong> Develop a system that parses factory service manuals
  based on the identified part to provide repair instructions alongside recognition results.
- <strong>Performance Metrics:</strong> Add analytics to track model performance and user interaction patterns.

## Contributing

Contributions are welcome! If you find any bugs or have suggestions for new features, feel free to
open an issue or submit a pull request.
