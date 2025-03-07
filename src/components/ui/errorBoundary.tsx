"use client";

import React, { Component, ErrorInfo, ReactNode } from "react";
import { Button } from "./button";

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
    };
  }

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error("Error caught by ErrorBoundary:", error, errorInfo);
  }

  resetError = (): void => {
    this.setState({
      hasError: false,
      error: null,
    });
  };

  render(): ReactNode {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="p-6 rounded-lg border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 text-center">
          <h2 className="text-xl font-semibold text-red-700 dark:text-red-400 mb-4">
            Something went wrong
          </h2>
          <p className="text-gray-700 dark:text-gray-300 mb-6">
            {this.state.error?.message || "An unexpected error occurred"}
          </p>
          <Button onClick={this.resetError} variant="outline">
            Try again
          </Button>
        </div>
      );
    }

    return this.props.children;
  }
} 