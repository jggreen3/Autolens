"use client";

import { useState, useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Camera, X, RefreshCw } from "lucide-react";

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
  const [facingMode, setFacingMode] = useState<'user' | 'environment'>('environment');

  useEffect(() => {
    let activeStream: MediaStream | null = null;
    let isMounted = true; // Prevents unnecessary updates

    navigator.mediaDevices
      .getUserMedia({ video: { facingMode: { ideal: facingMode } } })
      .then((mediaStream) => {
        if (!isMounted) return; // Prevent updates if component unmounted

        activeStream = mediaStream;
        setStream(mediaStream);
        console.log("MediaStream set:", mediaStream);

        if (videoRef.current) {
          videoRef.current.srcObject = mediaStream;

          videoRef.current.onloadedmetadata = () => {
            videoRef.current?.play().catch((error) => {
              console.error("Error trying to play video:", error);
            });
          };
        } else {
          console.warn("videoRef.current is not available yet.");
        }

        setIsWaitingPermission(false);
      })
      .catch((error) => {
        console.error("Camera error:", error);
        onClose();
      });

    return () => {
      isMounted = false; // Prevent state updates on unmounted component
      if (activeStream) {
        activeStream.getTracks().forEach((track) => track.stop());
      }
    };
  }, [facingMode]);

  useEffect(() => {
    if (videoRef.current && stream) {
      console.log("Assigning stream to video element");
      videoRef.current.srcObject = stream;
      videoRef.current.play().catch((error) => {
        console.error("Error trying to play video:", error);
      });
    }
  }, [stream]);

  const capturePhoto = () => {
    const video = videoRef.current;
    const canvas = document.createElement("canvas");

    if (video && stream) {
      canvas.width = video.videoWidth;
      canvas.height = video.videoHeight;
      const context = canvas.getContext("2d");
      context?.drawImage(video, 0, 0);

      stream.getTracks().forEach((track) => track.stop());
      setIsVideoActive(false);

      canvas.toBlob((blob) => {
        if (blob) {
          const file = new File([blob], "captured-photo.jpg", {
            type: "image/jpeg",
          });
          onCapture(file, URL.createObjectURL(blob));
          onClose();
        }
      }, "image/jpeg");
    }
  };

  const switchCamera = () => {
    if (stream) {
      stream.getTracks().forEach(track => track.stop());
      setFacingMode(prevMode => prevMode === 'user' ? 'environment' : 'user');
    }
  };

  const hasFrontAndBackCamera = () => {
    return 'mediaDevices' in navigator && 'enumerateDevices' in navigator.mediaDevices;
  };

  return (
    <div className="relative w-full h-[400px] bg-gray-100 dark:bg-auto-dark-card rounded-xl overflow-hidden shadow-auto">
      {isWaitingPermission ? (
        <div className="absolute inset-0 flex flex-col items-center justify-center space-y-4">
          <div className="p-4 bg-auto-blue/10 dark:bg-auto-blue/20 rounded-full animate-pulse">
            <Camera className="h-8 w-8 text-auto-blue dark:text-auto-blue-light" />
          </div>
          <p className="text-lg font-medium text-gray-700 dark:text-gray-200">
            Please allow camera access
          </p>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            We need permission to use your camera
          </p>
        </div>
      ) : (
        <>
          {isVideoActive && (
            <video
              ref={videoRef}
              autoPlay
              playsInline
              muted
              className="w-full h-full object-cover rounded-xl"
            />
          )}
          <div className="absolute bottom-4 left-0 right-0 flex justify-center gap-4 z-10">
            <Button
              onClick={capturePhoto}
              className="bg-auto-blue hover:bg-auto-blue-dark text-white button-glow"
            >
              <Camera className="mr-2 h-4 w-4" />
              Capture Photo
            </Button>
            
            {hasFrontAndBackCamera() && (
              <Button
                onClick={switchCamera}
                variant="outline"
                className="bg-white/80 dark:bg-gray-700/80 border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200"
              >
                <RefreshCw className="h-4 w-4" />
              </Button>
            )}
            
            <Button 
              variant="outline" 
              onClick={onClose}
              className="bg-white/80 dark:bg-gray-700/80 border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-200"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
