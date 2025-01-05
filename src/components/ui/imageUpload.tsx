'use client'

import { useRef } from 'react'
import { Button } from "@/components/ui/button"
import { Camera, Upload } from "lucide-react"

const ACCEPTED_IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/gif']

export function ImageUpload({ 
  onImageSelect,
  onCameraOpen
}: {
  onImageSelect: (file: File, preview: string) => void
  onCameraOpen: () => void
}) {
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (!ACCEPTED_IMAGE_TYPES.includes(file.type)) {
      alert('Please upload a JPG, PNG, or GIF image')
      e.target.value = ''
      return
    }

    onImageSelect(file, URL.createObjectURL(file))
  }

  return (
    <div className="flex gap-4">
      <Button 
        variant="outline" 
        className="flex-1"
        onClick={() => fileInputRef.current?.click()}
      >
        <Upload className="mr-2 h-4 w-4" />
        Upload Image
      </Button>
      
      <Button 
        variant="outline" 
        className="flex-1"
        onClick={onCameraOpen}
      >
        <Camera className="mr-2 h-4 w-4" />
        Take Photo
      </Button>

      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/png,image/gif"
        onChange={handleImageUpload}
        className="hidden"
      />
    </div>
  )
}
