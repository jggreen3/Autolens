import { Progress } from "@/components/ui/progress";

export type DetectionResultArray = [
  number, // index 0
  number, // index 1
  number, // index 2
  number, // index 3
  string, // index 4: label
  number // index 5: confidence
];

export function DetectionResults({
  results,
}: {
  results: DetectionResultArray[];
}) {
  const formatConfidence = (confidence: number): number => {
    return Math.round(confidence * 100);
  };

  return (
    <div className="w-full max-w-xl mt-8">
      <h2 className="text-xl font-semibold mb-4 dark:text-gray-100">
        Detection Results:
      </h2>
      {results.length === 0 ? (
        <div className="text-center text-gray-500 dark:text-gray-300">
          No objects detected. Please try again with a different picture.
        </div>
      ) : (
        <div className="space-y-4">
          {results.map((result, index) => (
            <div
              key={index}
              className="bg-white dark:bg-dark-background p-4 rounded-lg border dark:border-gray-700"
            >
              <div className="flex justify-between items-center mb-2">
                <span className="font-medium dark:text-gray-100">
                  {result[4]}
                </span>
                <span className="text-sm text-muted-foreground dark:text-gray-300">
                  {formatConfidence(result[5])}% confidence
                </span>
              </div>
              <Progress value={formatConfidence(result[5])} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
