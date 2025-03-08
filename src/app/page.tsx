"use client";

import { useState } from "react";
import { ImageUpload } from "@/components/ui/imageUpload";
import { CameraCapture } from "@/components/ui/cameraCapture";
import { DetectionResults } from "@/components/ui/detectionResults";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { ErrorBoundary } from "@/components/ui/errorBoundary";
import Image from "next/image";
import type { DetectionResultArray } from "@/components/ui/detectionResults";
import useSystemDarkMode from "@/hooks/useSystemDarkMode";

// Loading skeleton component for results
const ResultsSkeleton = () => (
  <div className="w-full max-w-xl mt-8 space-y-4">
    <div className="h-7 w-48 bg-gray-200 dark:bg-gray-700 rounded-md animate-pulse mb-4"></div>
    {[1, 2, 3].map((i) => (
      <div key={i} className="bg-white dark:bg-gray-800 p-4 rounded-lg border dark:border-gray-700">
        <div className="flex justify-between items-center mb-2">
          <div className="h-6 w-32 bg-gray-200 dark:bg-gray-700 rounded-md animate-pulse"></div>
          <div className="h-5 w-24 bg-gray-200 dark:bg-gray-700 rounded-md animate-pulse"></div>
        </div>
        <div className="h-2 bg-gray-200 dark:bg-gray-700 rounded-full animate-pulse"></div>
      </div>
    ))}
  </div>
);

export default function Home() {
  const [selectedImage, setSelectedImage] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string>("");
  const [results, setResults] = useState<DetectionResultArray[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [showCamera, setShowCamera] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useSystemDarkMode();

  const handleImageSelect = (file: File, preview: string) => {
    setSelectedImage(file);
    setImagePreview(preview);
  };

  const handleSubmit = async () => {
    if (!selectedImage) return;

    setLoading(true);
    setError(null);
    const formData = new FormData();
    formData.append("image_file", selectedImage);

    try {
      const response = await fetch("/api/detect", {
        method: "POST",
        body: formData,
      });
      
      if (!response.ok) {
        throw new Error(`Server responded with status: ${response.status}`);
      }
      
      const data = await response.json();
      setResults(data);
    } catch (error) {
      console.error("Error:", error);
      setError("Failed to process image. Please try again or use a different image.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center bg-gradient-to-b from-background to-secondary dark:from-auto-black dark:to-auto-gray">
      {/* Hero Section */}
      <div className="w-full py-12 md:py-20 px-4 sm:px-6 md:px-12 bg-gradient-auto from-auto-blue-light to-auto-blue dark:from-auto-blue-dark dark:to-auto-blue">
        <div className="max-w-4xl mx-auto text-center space-y-6">
          <h1 className="text-4xl md:text-6xl font-bold text-white text-shadow">
            AutoLens
          </h1>
          <p className="text-xl md:text-2xl text-white/90">
            Discover and Identify Car Parts with AI
          </p>
          <p className="text-base md:text-lg text-white/80 max-w-2xl mx-auto">
            Our AI technology can recognize car parts from any model instantly. 
            Simply upload an image you need to identify, and let AutoLens do the work for you.
          </p>
        </div>
      </div>

      {/* Main Content */}
      <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 -mt-8 md:-mt-12 mb-20 flex-grow">
        <div className="bg-white dark:bg-auto-dark-card rounded-xl shadow-auto p-6 md:p-8">
          <ErrorBoundary>
            <div className="space-y-8">
              <div className="space-y-4">
                <h2 className="text-2xl font-semibold text-auto-blue-dark dark:text-auto-silver">
                  Upload Your Image
                </h2>
                <p className="text-muted-foreground dark:text-gray-400">
                  Take a photo or upload an image of the car part you want to identify
                </p>
                <ImageUpload
                  onImageSelect={handleImageSelect}
                  onCameraOpen={() => setShowCamera(true)}
                />
              </div>

              {showCamera && (
                <div className="mt-6">
                  <CameraCapture
                    onCapture={handleImageSelect}
                    onClose={() => setShowCamera(false)}
                  />
                </div>
              )}

              {imagePreview && (
                <div className="space-y-4">
                  <h3 className="text-lg font-medium text-gray-900 dark:text-gray-200">
                    Preview
                  </h3>
                  <div className="relative w-full aspect-[16/10] rounded-lg overflow-hidden shadow-md">
                    <Image
                      src={imagePreview}
                      alt="Preview"
                      fill
                      className="object-contain"
                      priority
                    />
                  </div>
                </div>
              )}

              {selectedImage && (
                <Button
                  onClick={handleSubmit}
                  disabled={loading}
                  className="w-full bg-auto-blue hover:bg-auto-blue-dark text-white py-6 text-lg rounded-lg button-glow"
                >
                  {loading ? "Processing..." : "Detect Objects"}
                </Button>
              )}

              {loading && (
                <div className="mt-4 space-y-2">
                  <div className="flex justify-between text-sm text-muted-foreground dark:text-gray-400">
                    <span>Processing image...</span>
                    <span>Please wait</span>
                  </div>
                  <Progress value={100} className="animate-pulse bg-auto-silver h-2" />
                </div>
              )}

              {error && (
                <div className="p-4 mt-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md text-red-600 dark:text-red-400">
                  {error}
                </div>
              )}

              {loading ? <ResultsSkeleton /> : results && <DetectionResults results={results} />}
            </div>
          </ErrorBoundary>
        </div>
      </div>

      {/* Footer */}
      <footer className="w-full py-6 bg-auto-blue-dark text-white/80 text-center text-sm mt-auto">
        <div className="max-w-4xl mx-auto px-4">
          <p>AutoLens — Powered by YOLOv8 and Next.js</p>
        </div>
      </footer>
    </div>
  );
}
