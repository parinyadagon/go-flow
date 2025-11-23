"use client";

import React from "react";
import { AlertTriangle, RefreshCw } from "lucide-react";

interface ErrorBoundaryProps {
  children: React.ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error("ErrorBoundary caught an error:", error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.href = "/";
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-slate-900 p-4">
          <div className="max-w-md w-full">
            <div className="bg-white dark:bg-slate-800 rounded-xl shadow-xl p-8 border border-red-200 dark:border-red-500/30">
              <div className="flex items-center justify-center w-16 h-16 rounded-full bg-red-100 dark:bg-red-500/10 mx-auto mb-4">
                <AlertTriangle className="w-8 h-8 text-red-600 dark:text-red-400" />
              </div>
              <h2 className="text-2xl font-bold text-center text-gray-900 dark:text-white mb-2">Oops! Something went wrong</h2>
              <p className="text-center text-gray-600 dark:text-gray-400 mb-4">
                The application encountered an unexpected error. Don&apos;t worry, this has been logged and we&apos;ll fix it!
              </p>
              {this.state.error && (
                <div className="bg-red-50 dark:bg-red-500/5 rounded-lg p-4 mb-6">
                  <p className="text-xs font-mono text-red-800 dark:text-red-400 break-all">{this.state.error.message}</p>
                </div>
              )}
              <div className="flex gap-3">
                <button
                  onClick={this.handleReset}
                  className="flex-1 px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 transition-all duration-200 flex items-center justify-center gap-2 font-medium shadow-md hover:shadow-lg">
                  <RefreshCw className="w-4 h-4" />
                  Go to Home
                </button>
                <button
                  onClick={() => window.location.reload()}
                  className="flex-1 px-4 py-2 bg-gray-200 dark:bg-slate-700 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-slate-600 transition-all duration-200 font-medium">
                  Reload Page
                </button>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
