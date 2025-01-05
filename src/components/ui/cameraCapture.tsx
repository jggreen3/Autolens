"use client";

import { useState, useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Camera, X } from "lucide-react";

export function CameraCapture({
  onCapture,
  onClose,
}: {
  onCapture: (file: File, preview: string) => void;
  onClose: () => void;
}) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [isWaitingPermission, setIsWaitingPermission] = useState(true);
  const [stream, setStream] = useState<MediaStream | null>(null);
  const [isVideoActive, setIsVideoActive] = useState(true);

  useEffect(() => {
    navigator.mediaDevices
      .getUserMedia({ video: true })
      .then((mediaStream) => {
        setStream(mediaStream);
        if (videoRef.current) {
          videoRef.current.srcObject = mediaStream;
        }
        setIsWaitingPermission(false);
      })
      .catch((error) => {
        console.log("Camera error:", error);
        onClose();
      });

    return () => {
      if (stream) {
        stream.getTracks().forEach((track) => track.stop());
      }
    };
  }, []);

  const capturePhoto = () => {
    const video = videoRef.current
    const canvas = document.createElement('canvas')
    
    if (video && stream) {
      canvas.width = video.videoWidth
      canvas.height = video.videoHeight
      const context = canvas.getContext('2d')
      context?.drawImage(video, 0, 0)
      
      stream.getTracks().forEach(track => track.stop())
      setIsVideoActive(false)
      
      canvas.toBlob((blob) => {
        if (blob) {
          const file = new File([blob], "captured-photo.jpg", { type: "image/jpeg" })
          onCapture(file, URL.createObjectURL(blob))
          onClose() // Close the camera component after capture
        }
      }, 'image/jpeg')
    }
  };

  return (
    <div className="relative w-full h-[400px] bg-slate-100 rounded-lg flex items-center justify-center">
      {isWaitingPermission ? (
        <div className="text-center">
          <Camera className="h-8 w-8 mx-auto mb-2" />
          <p>Please allow camera access</p>
        </div>
      ) : (
        <>
          {isVideoActive && (
            <video
              ref={videoRef}
              autoPlay
              playsInline
              muted
              className="w-full h-full object-cover rounded-lg"
            />
          )}
          <div className="absolute bottom-4 left-0 right-0 flex justify-center gap-4 z-10">
            <Button
              onClick={capturePhoto}
              className="bg-green-500 hover:bg-green-600"
            >
              Capture Photo
            </Button>
            <Button variant="destructive" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
