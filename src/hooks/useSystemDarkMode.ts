import { useEffect } from "react";

function useSystemDarkMode() {
  useEffect(() => {
    const root = window.document.documentElement;
    const darkModeMediaQuery = window.matchMedia(
      "(prefers-color-scheme: dark)"
    );

    const applyDarkMode = (isDarkMode: boolean) => {
      if (isDarkMode) {
        root.classList.add("dark");
      } else {
        root.classList.remove("dark");
      }
    };

    // Apply the initial preference
    applyDarkMode(darkModeMediaQuery.matches);

    // Listen for changes in the preference
    const handleChange = (event: MediaQueryListEvent) => {
      applyDarkMode(event.matches);
    };

    darkModeMediaQuery.addEventListener("change", handleChange);

    // Cleanup listener on unmount
    return () => {
      darkModeMediaQuery.removeEventListener("change", handleChange);
    };
  }, []);
}

export default useSystemDarkMode;
