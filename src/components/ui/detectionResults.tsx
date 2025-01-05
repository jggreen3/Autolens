import { Progress } from "@/components/ui/progress"

export function DetectionResults({ results }: { results: any[] }) {
    const formatConfidence = (confidence: number): number => {
      return Math.round(confidence * 100)
    }
  
    return (
      <div className="w-full max-w-xl mt-8">
        <h2 className="text-xl font-semibold mb-4">Detection Results:</h2>
        <div className="space-y-4">
          {results.map((result, index) => (
            <div key={index} className="bg-white p-4 rounded-lg border">
              <div className="flex justify-between items-center mb-2">
                <span className="font-medium">{result[4]}</span>
                <span className="text-sm text-muted-foreground">
                  {formatConfidence(result[5])}% confidence
                </span>
              </div>
              <Progress value={formatConfidence(result[5])} />
            </div>
          ))}
        </div>
      </div>
    )
  }
  