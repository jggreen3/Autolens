"use client";

import { useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Camera, Upload, Image as ImageIcon } from "lucide-react";

const ACCEPTED_IMAGE_TYPES = ["image/jpeg", "image/png", "image/gif"];
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

export function ImageUpload({
  onImageSelect,
  onCameraOpen,
}: {
  onImageSelect: (file: File, preview: string) => void;
  onCameraOpen: () => void;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    setError(null);
    
    if (!file) return;

    if (!ACCEPTED_IMAGE_TYPES.includes(file.type)) {
      setError("Please upload a JPG, PNG, or GIF image");
      e.target.value = "";
      return;
    }

    if (file.size > MAX_FILE_SIZE) {
      setError("File size exceeds 10MB limit");
      e.target.value = "";
      return;
    }

    // Create a preview URL for the image
    const preview = URL.createObjectURL(file);
    
    // Optionally resize large images before uploading
    if (file.size > 2 * 1024 * 1024) { // If larger than 2MB
      resizeImage(file, preview, (resizedFile, resizedPreview) => {
        onImageSelect(resizedFile, resizedPreview);
      });
    } else {
      onImageSelect(file, preview);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(false);
    setError(null);
    
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const file = e.dataTransfer.files[0];
      
      if (!ACCEPTED_IMAGE_TYPES.includes(file.type)) {
        setError("Please upload a JPG, PNG, or GIF image");
        return;
      }

      if (file.size > MAX_FILE_SIZE) {
        setError("File size exceeds 10MB limit");
        return;
      }

      const preview = URL.createObjectURL(file);
      
      if (file.size > 2 * 1024 * 1024) {
        resizeImage(file, preview, (resizedFile, resizedPreview) => {
          onImageSelect(resizedFile, resizedPreview);
        });
      } else {
        onImageSelect(file, preview);
      }
    }
  };

  const resizeImage = (
    file: File, 
    originalPreview: string, 
    callback: (resizedFile: File, resizedPreview: string) => void
  ) => {
    const img = new Image();
    img.src = originalPreview;
    
    img.onload = () => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      
      // Calculate new dimensions (max 1200px width/height)
      const MAX_DIMENSION = 1200;
      let width = img.width;
      let height = img.height;
      
      if (width > height && width > MAX_DIMENSION) {
        height = Math.round(height * (MAX_DIMENSION / width));
        width = MAX_DIMENSION;
      } else if (height > MAX_DIMENSION) {
        width = Math.round(width * (MAX_DIMENSION / height));
        height = MAX_DIMENSION;
      }
      
      canvas.width = width;
      canvas.height = height;
      
      ctx?.drawImage(img, 0, 0, width, height);
      
      // Convert to blob and create a new File
      canvas.toBlob((blob) => {
        if (blob) {
          const resizedFile = new File([blob], file.name, {
            type: 'image/jpeg',
            lastModified: Date.now(),
          });
          
          const resizedPreview = URL.createObjectURL(blob);
          callback(resizedFile, resizedPreview);
        } else {
          // Fallback to original if resize fails
          callback(file, originalPreview);
        }
      }, 'image/jpeg', 0.85); // 85% quality JPEG
    };
    
    img.onerror = () => {
      // Fallback to original if loading fails
      callback(file, originalPreview);
    };
  };

  return (
    <div className="space-y-4">
      <div 
        className={`border-2 border-dashed rounded-xl p-8 text-center transition-colors ${
          isDragging 
            ? "border-auto-blue bg-auto-blue/5 dark:border-auto-blue-light dark:bg-auto-blue/10" 
            : "border-gray-300 dark:border-gray-600 hover:border-auto-blue dark:hover:border-auto-blue-light"
        }`}
        onDragOver={(e) => {
          e.preventDefault();
          setIsDragging(true);
        }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={handleDrop}
      >
        <div className="flex flex-col items-center justify-center space-y-4">
          <div className="p-3 bg-auto-blue/10 dark:bg-auto-blue/20 rounded-full">
            <ImageIcon className="h-8 w-8 text-auto-blue dark:text-auto-blue-light" />
          </div>
          <div>
            <p className="text-lg font-medium text-gray-700 dark:text-gray-200">
              Drag and drop your image here
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              or select an option below
            </p>
          </div>
        </div>
      </div>

      <div className="flex gap-4" role="group" aria-label="Image upload options">
        <Button
          variant="outline"
          className="flex-1 border-auto-blue text-auto-blue hover:bg-auto-blue/10 dark:border-auto-blue-light dark:text-auto-blue-light dark:hover:bg-auto-blue/20"
          onClick={() => fileInputRef.current?.click()}
          aria-label="Upload image from device"
        >
          <Upload className="mr-2 h-4 w-4" aria-hidden="true" />
          Upload Image
        </Button>

        <Button
          variant="outline"
          className="flex-1 border-auto-blue text-auto-blue hover:bg-auto-blue/10 dark:border-auto-blue-light dark:text-auto-blue-light dark:hover:bg-auto-blue/20"
          onClick={onCameraOpen}
          aria-label="Take photo with camera"
        >
          <Camera className="mr-2 h-4 w-4" aria-hidden="true" />
          Take Photo
        </Button>

        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/png,image/gif"
          onChange={handleImageUpload}
          className="hidden"
          aria-label="File upload"
        />
      </div>
      
      {error && (
        <div className="text-red-500 text-sm mt-2 p-3 bg-red-50 dark:bg-red-900/10 rounded-lg border border-red-200 dark:border-red-800">
          {error}
        </div>
      )}
    </div>
  );
}
