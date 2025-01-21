"use client";

import { useState } from "react";
import { ImageUpload } from "@/components/ui/imageUpload";
import { CameraCapture } from "@/components/ui/cameraCapture";
import { DetectionResults } from "@/components/ui/detectionResults";
import { Button } from "@/components/ui/button";
import Image from "next/image";
import type { DetectionResultArray } from "@/components/ui/detectionResults";

export default function Home() {
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string>("");
  const [results, setResults] = useState<DetectionResultArray[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [showCamera, setShowCamera] = useState(false);

  const handleImageSelect = (file: File, preview: string) => {
    setSelectedImage(file);
    setImagePreview(preview);
  };

  const handleSubmit = async () => {
    if (!selectedImage) return;

    setLoading(true);
    const formData = new FormData();
    formData.append("image_file", selectedImage);

    try {
      const response = await fetch("/api/detect", {
        method: "POST",
        body: formData,
      });
      const data = await response.json();
      setResults(data);
    } catch (error) {
      console.error("Error:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen pt-8 p-12 flex flex-col items-center">
      <div className="max-w-2xl text-center space-y-8 mb-12">
        <h1 className="text-5xl font-bold">AutoLens</h1>
        <div className="space-y-6">
          <p className="text-2xl text-muted-foreground">
            Discover and Identify Car Parts with AI
          </p>
          <p className="text-lg text-muted-foreground">
            Welcome to your personal mechanic. Our AI technology can recognize
            car parts from any model instantly. Simply upload an image you need
            to identify, and let AutoLens do the work for you.
          </p>
        </div>
      </div>

      <div className="w-full max-w-xl space-y-6">
        <ImageUpload
          onImageSelect={handleImageSelect}
          onCameraOpen={() => setShowCamera(true)}
        />

        {showCamera && (
          <CameraCapture
            onCapture={handleImageSelect}
            onClose={() => setShowCamera(false)}
          />
        )}

        {imagePreview && (
          <div className="relative w-full max-w-xl aspect-[16/10]">
            <Image
              src={imagePreview}
              alt="Preview"
              fill
              className="object-contain rounded-lg border"
            />
          </div>
        )}

        {selectedImage && (
          <Button onClick={handleSubmit} disabled={loading} className="w-full">
            {loading ? "Processing..." : "Detect Objects"}
          </Button>
        )}

        {results && <DetectionResults results={results} />}
      </div>
    </div>
  );
}
