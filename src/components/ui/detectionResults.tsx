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

  // Sort results by confidence (highest first)
  const sortedResults = [...results].sort((a, b) => b[5] - a[5]);

  // Function to determine confidence level class
  const getConfidenceClass = (confidence: number): string => {
    const percent = formatConfidence(confidence);
    if (percent >= 90) return "text-green-600 dark:text-green-400";
    if (percent >= 70) return "text-auto-blue dark:text-auto-blue-light";
    return "text-amber-600 dark:text-amber-400";
  };

  // Function to determine progress bar color
  const getProgressColorClass = (confidence: number): string => {
    const percent = formatConfidence(confidence);
    if (percent >= 90) return "bg-green-500 dark:bg-green-400";
    if (percent >= 70) return "bg-auto-blue dark:bg-auto-blue-light";
    return "bg-amber-500 dark:bg-amber-400";
  };

  return (
    <div className="w-full mt-8 space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold text-auto-blue-dark dark:text-auto-blue-light">
          Detection Results
        </h2>
        <span className="text-sm text-muted-foreground dark:text-gray-300">
          {sortedResults.length} {sortedResults.length === 1 ? 'item' : 'items'} detected
        </span>
      </div>
      
      {sortedResults.length === 0 ? (
        <div className="text-center p-8 bg-white dark:bg-auto-dark-card rounded-xl border dark:border-gray-600 shadow-auto">
          <div className="text-gray-500 dark:text-gray-300 mb-3 text-lg">
            No objects detected
          </div>
          <p className="text-sm text-gray-400 dark:text-gray-400 max-w-md mx-auto">
            Please try again with a different picture or adjust the angle/lighting for better results
          </p>
        </div>
      ) : (
        <div className="space-y-4 max-w-3xl mx-auto">
          {sortedResults.map((result, index) => (
            <div
              key={index}
              className="bg-white dark:bg-auto-dark-card p-5 rounded-xl border dark:border-gray-600 transition-all hover:shadow-auto-hover card-hover"
            >
              <div className="flex justify-between items-center mb-3">
                <span className="font-medium text-lg text-gray-900 dark:text-gray-100">
                  {result[4]}
                </span>
                <span 
                  className={`text-sm font-medium px-3 py-1 rounded-full ${
                    formatConfidence(result[5]) >= 90 
                      ? "bg-green-100 dark:bg-green-900/30" 
                      : formatConfidence(result[5]) >= 70
                        ? "bg-blue-100 dark:bg-blue-900/30"
                        : "bg-amber-100 dark:bg-amber-900/30"
                  } ${getConfidenceClass(result[5])}`}
                >
                  {formatConfidence(result[5])}%
                </span>
              </div>
              <div className="relative pt-1">
                <div className="flex mb-2 items-center justify-between">
                  <div>
                    <span className="text-xs font-semibold inline-block text-gray-600 dark:text-gray-300">
                      Confidence
                    </span>
                  </div>
                </div>
                <div className="overflow-hidden h-2 mb-1 text-xs flex rounded-full bg-gray-200 dark:bg-gray-700">
                  <div 
                    style={{ width: `${formatConfidence(result[5])}%` }} 
                    className={`shadow-none flex flex-col text-center whitespace-nowrap text-white justify-center transition-all duration-500 ${getProgressColorClass(result[5])}`}
                  ></div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

