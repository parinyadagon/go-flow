"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus, Eye, RefreshCw, Loader2, AlertCircle, ChevronLeft, ChevronRight } from "lucide-react";

interface Workflow {
  ID: string;
  WorkflowName: string;
  Status: string;
  CreatedAt: string;
  UpdatedAt: string;
}

interface WorkflowsResponse {
  workflows: Workflow[];
  limit: number;
  offset: number;
  total: number;
  totalPages: number;
  currentPage: number;
}

export default function Home() {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [itemsPerPage] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);

  const fetchWorkflows = (page: number) => {
    setLoading(true);
    setError(null);
    const offset = (page - 1) * itemsPerPage;

    fetch(`http://localhost:8080/workflows?limit=${itemsPerPage}&offset=${offset}`)
      .then((res) => {
        if (!res.ok) {
          throw new Error(`HTTP error! status: ${res.status}`);
        }
        return res.json();
      })
      .then((data: WorkflowsResponse) => {
        console.log("API Response:", data);
        setWorkflows(data.workflows || []);
        setTotal(data.total || 0);
        setTotalPages(data.totalPages || 1);
        setLoading(false);
      })
      .catch((err) => {
        console.error("Fetch error:", err);
        setError(err.message);
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchWorkflows(currentPage);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentPage]);

  const handlePreviousPage = () => {
    if (currentPage > 1) {
      setCurrentPage(currentPage - 1);
    }
  };

  const handleNextPage = () => {
    if (currentPage < totalPages) {
      setCurrentPage(currentPage + 1);
    }
  };

  const refreshData = () => {
    fetchWorkflows(currentPage);
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 dark:bg-slate-900">
        <div className="text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-blue-100 dark:bg-blue-500/10 mb-4">
            <Loader2 className="w-8 h-8 text-blue-600 dark:text-blue-400 animate-spin" />
          </div>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">Loading Workflows</h2>
          <p className="text-sm text-gray-600 dark:text-gray-400">Please wait while we fetch your data...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 dark:bg-slate-900 p-4">
        <div className="max-w-md w-full">
          <div className="bg-white dark:bg-slate-800 rounded-xl shadow-xl p-8 border border-red-200 dark:border-red-500/30">
            <div className="flex items-center justify-center w-16 h-16 rounded-full bg-red-100 dark:bg-red-500/10 mx-auto mb-4">
              <AlertCircle className="w-8 h-8 text-red-600 dark:text-red-400" />
            </div>
            <h2 className="text-2xl font-bold text-center text-gray-900 dark:text-white mb-2">Oops! Something went wrong</h2>
            <p className="text-center text-gray-600 dark:text-gray-400 mb-6">
              We couldn&apos;t load the workflows. Please check if the server is running.
            </p>
            <div className="bg-red-50 dark:bg-red-500/5 rounded-lg p-4 mb-6">
              <p className="text-sm font-mono text-red-800 dark:text-red-400">Error: {error}</p>
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => {
                  setError(null);
                  fetchWorkflows(currentPage);
                }}
                className="flex-1 px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 transition-all duration-200 flex items-center justify-center gap-2 font-medium">
                <RefreshCw className="w-4 h-4" />
                Try Again
              </button>
              <button
                onClick={() => setError(null)}
                className="flex-1 px-4 py-2 bg-gray-200 dark:bg-slate-700 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-slate-600 transition-all duration-200 font-medium">
                Dismiss
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-slate-900 text-gray-900 dark:text-white transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">🚀 Workflow Instances</h1>
            <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">View and manage all workflow instances</p>
          </div>
          <div className="flex gap-3">
            <button
              onClick={refreshData}
              className="px-4 py-2 bg-gray-200 dark:bg-slate-700 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-slate-600 transition-all duration-200 flex items-center gap-2 font-medium shadow-sm hover:shadow-md">
              <RefreshCw className="w-4 h-4" />
              Refresh
            </button>
            <button className="px-5 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 transition-all duration-200 flex items-center gap-2 font-medium shadow-md hover:shadow-lg transform hover:scale-105">
              <Plus className="w-5 h-5" />
              Create Workflow
            </button>
          </div>
        </div>

        <div className="bg-white dark:bg-slate-800 shadow-xl overflow-hidden rounded-xl border border-gray-200 dark:border-slate-700 transition-colors">
          <table className="min-w-full divide-y divide-gray-200 dark:divide-slate-700">
            <thead className="bg-gray-50 dark:bg-slate-700/50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">ID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Workflow Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Created At</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="bg-white dark:bg-slate-800 divide-y divide-gray-200 dark:divide-slate-700">
              {workflows.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-4 text-center text-sm text-gray-500 dark:text-gray-400">
                    No workflows found
                  </td>
                </tr>
              ) : (
                workflows.map((workflow, index) => (
                  <tr key={workflow.ID || `workflow-${index}`} className="hover:bg-gray-50 dark:hover:bg-slate-700/30 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-700 dark:text-gray-300">{workflow.ID || "N/A"}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-semibold text-gray-900 dark:text-white">
                      {workflow.WorkflowName || "N/A"}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span
                        className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                          workflow.Status === "COMPLETED"
                            ? "bg-green-100 dark:bg-green-500/10 text-green-800 dark:text-green-400 border border-green-500"
                            : workflow.Status === "FAILED"
                            ? "bg-red-100 dark:bg-red-500/10 text-red-800 dark:text-red-400 border border-red-500"
                            : workflow.Status === "RUNNING"
                            ? "bg-yellow-100 dark:bg-yellow-400/10 text-yellow-800 dark:text-yellow-400 border border-yellow-400"
                            : "bg-gray-100 dark:bg-slate-600/50 text-gray-800 dark:text-slate-400 border border-gray-300 dark:border-slate-600"
                        }`}>
                        {workflow.Status || "UNKNOWN"}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                      {workflow.CreatedAt ? new Date(workflow.CreatedAt).toLocaleString() : "N/A"}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                      <Link
                        href={`/workflows/${workflow.ID}`}
                        className="inline-flex items-center gap-2 px-4 py-2 bg-blue-50 dark:bg-blue-500/10 text-blue-600 dark:text-blue-400 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-500/20 border border-blue-200 dark:border-blue-500/30 transition-all duration-200 font-medium hover:shadow-md group">
                        <Eye className="w-4 h-4" />
                        <span>View Details</span>
                        <span className="transform group-hover:translate-x-1 transition-transform duration-200">→</span>
                      </Link>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>

          {/* Pagination Controls */}
          <div className="bg-gray-50 dark:bg-slate-700/50 px-6 py-4 border-t border-gray-200 dark:border-slate-700">
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-600 dark:text-gray-400">
                Page <span className="font-semibold text-gray-900 dark:text-white">{currentPage}</span> of{" "}
                <span className="font-semibold text-gray-900 dark:text-white">{totalPages}</span>
                {total > 0 && (
                  <span className="ml-2">
                    (Total: {total} {total === 1 ? "workflow" : "workflows"})
                  </span>
                )}
              </div>
              <div className="flex gap-2">
                <button
                  onClick={handlePreviousPage}
                  disabled={currentPage === 1}
                  className={`px-4 py-2 rounded-lg font-medium transition-all duration-200 flex items-center gap-2 ${
                    currentPage === 1
                      ? "bg-gray-100 dark:bg-slate-800 text-gray-400 dark:text-slate-600 cursor-not-allowed"
                      : "bg-white dark:bg-slate-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-slate-600 hover:bg-gray-50 dark:hover:bg-slate-700 shadow-sm hover:shadow-md"
                  }`}>
                  <ChevronLeft className="w-4 h-4" />
                  Previous
                </button>
                <button
                  onClick={handleNextPage}
                  disabled={currentPage >= totalPages}
                  className={`px-4 py-2 rounded-lg font-medium transition-all duration-200 flex items-center gap-2 ${
                    currentPage >= totalPages
                      ? "bg-gray-100 dark:bg-slate-800 text-gray-400 dark:text-slate-600 cursor-not-allowed"
                      : "bg-white dark:bg-slate-800 text-gray-700 dark:text-gray-200 border border-gray-300 dark:border-slate-600 hover:bg-gray-50 dark:hover:bg-slate-700 shadow-sm hover:shadow-md"
                  }`}>
                  Next
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
