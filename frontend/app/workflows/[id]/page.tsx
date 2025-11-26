"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import useSWR from "swr";
import axios from "axios";
import { CheckCircle, Circle, Clock, AlertTriangle, ArrowLeft, RefreshCw, Loader2, AlertCircle } from "lucide-react";
import { useState } from "react";

interface Task {
  ID: number;
  TaskName: string;
  Status: string;
  RetryCount?: number;
  MaxRetries?: number;
  UpdatedAt?: string;
}

interface ActivityLog {
  ID: number;
  WorkflowInstanceID: string;
  TaskName?: string;
  EventType: string;
  Details: string;
  CreatedAt: string;
}

interface WorkflowData {
  workflow: {
    ID: string;
    WorkflowName: string;
    Status: string;
  };
  tasks: Task[];
  activityLogs: ActivityLog[];
}

// Fetcher function สำหรับ SWR
const fetcher = (url: string) => axios.get(url).then((res) => res.data);

export default function WorkflowDetail() {
  const params = useParams();
  const workflowId = params.id as string;
  const [isRefreshing, setIsRefreshing] = useState(false);
  // Removed retryingTaskId state

  // 🔥 Magic: refreshInterval จะยิง API ทุก 1 วินาทีเพื่ออัปเดตสถานะ Real-time!
  const { data, error, mutate } = useSWR<WorkflowData>(workflowId ? `http://localhost:8080/workflows/${workflowId}` : null, fetcher, {
    refreshInterval: 1000,
  });

  const handleRefresh = async () => {
    setIsRefreshing(true);
    await mutate();
    setTimeout(() => setIsRefreshing(false), 500);
  };

  // Removed handleRetryTask function

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 dark:bg-slate-900 p-4">
        <div className="max-w-md w-full">
          <div className="bg-white dark:bg-slate-800 rounded-xl shadow-xl p-8 border border-red-200 dark:border-red-500/30">
            <div className="flex items-center justify-center w-16 h-16 rounded-full bg-red-100 dark:bg-red-500/10 mx-auto mb-4">
              <AlertCircle className="w-8 h-8 text-red-600 dark:text-red-400" />
            </div>
            <h2 className="text-2xl font-bold text-center text-gray-900 dark:text-white mb-2">Failed to Load Workflow</h2>
            <p className="text-center text-gray-600 dark:text-gray-400 mb-6">
              We couldn&apos;t load the workflow details. The workflow might not exist or the server is unavailable.
            </p>
            <div className="flex gap-3">
              <Link
                href="/"
                className="flex-1 px-4 py-2 bg-gray-200 dark:bg-slate-700 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-slate-600 transition-all duration-200 flex items-center justify-center gap-2 font-medium">
                <ArrowLeft className="w-4 h-4" />
                Back to List
              </Link>
              <button
                onClick={handleRefresh}
                disabled={isRefreshing}
                className="flex-1 px-4 py-2 bg-blue-600 dark:bg-blue-700 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-800 transition-all duration-200 flex items-center justify-center gap-2 font-medium disabled:opacity-50 disabled:cursor-not-allowed">
                <RefreshCw className={`w-4 h-4 ${isRefreshing ? "animate-spin" : ""}`} />
                Try Again
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 dark:bg-slate-900">
        <div className="text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-blue-100 dark:bg-blue-500/10 mb-4">
            <Loader2 className="w-8 h-8 text-blue-600 dark:text-blue-400 animate-spin" />
          </div>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">Loading Workflow</h2>
          <p className="text-sm text-gray-600 dark:text-gray-400">Fetching workflow details...</p>
        </div>
      </div>
    );
  }

  const { workflow, tasks, activityLogs } = data;

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-slate-900 text-gray-900 dark:text-white transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header with Back Button */}
        <div className="mb-8">
          <div className="flex justify-between items-center">
            <div>
              <Link
                href="/"
                className="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-slate-800 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-200 dark:hover:bg-slate-700 border border-gray-300 dark:border-slate-600 transition-all duration-200 font-medium group mb-4">
                <ArrowLeft className="w-4 h-4 transform group-hover:-translate-x-1 transition-transform duration-200" />
                Back to Workflows
              </Link>
            </div>
            <div className="flex items-center gap-3">
              <div className="text-gray-500 dark:text-gray-400 text-sm font-mono">{workflow.ID}</div>
              <button
                onClick={handleRefresh}
                disabled={isRefreshing}
                className="px-4 py-2 bg-gray-100 dark:bg-slate-800 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-200 dark:hover:bg-slate-700 border border-gray-300 dark:border-slate-600 transition-all duration-200 flex items-center gap-2 font-medium disabled:opacity-50 disabled:cursor-not-allowed">
                <RefreshCw className={`w-4 h-4 ${isRefreshing ? "animate-spin" : ""}`} />
                Refresh
              </button>
            </div>
          </div>
          <div className="flex justify-between items-center">
            <h1 className="text-3xl font-bold flex items-center gap-2">
              🚀 {workflow.WorkflowName}
              <span className="text-sm bg-blue-600 px-2 py-1 rounded-full text-white">{workflow.Status}</span>
            </h1>
          </div>
        </div>

        {/* Visualization Area */}
        <div className="bg-white dark:bg-slate-800 p-8 rounded-xl shadow-2xl overflow-x-auto transition-colors">
          <div className="flex items-center gap-6 min-w-max py-4">
            {/* Loop Render Tasks */}
            {tasks.map((task: Task, index: number) => (
              <div key={task.ID} className="flex items-center">
                {/* Compact Task Card */}
                <div className="relative group">
                  <div
                    className={`
                    relative w-44 rounded-lg border-2 overflow-hidden transition-all duration-300
                    hover:scale-105 hover:-translate-y-1 cursor-pointer
                    ${getStatusStyle(task.Status)}
                  `}>
                    {/* Colored Header Bar */}
                    <div className={`h-1.5 ${getHeaderColor(task.Status)}`}></div>

                    {/* Card Content */}
                    <div className="p-3">
                      {/* Task Name with Icon */}
                      <div className="flex items-center gap-2 mb-2">
                        <div className={`p-1.5 rounded-lg ${getIconBg(task.Status)}`}>{getIcon(task.Status, 20)}</div>
                        <h3 className="font-semibold text-sm text-gray-800 dark:text-white leading-tight flex-1 truncate">{task.TaskName}</h3>
                      </div>

                      {/* Status Badge with Retry Count */}
                      <div className="flex items-center justify-between gap-2">
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${getStatusBadge(task.Status)}`}>
                          <span
                            className={`w-1 h-1 rounded-full mr-1 ${task.Status === "RUNNING" ? "animate-pulse" : ""} ${getStatusDot(
                              task.Status
                            )}`}></span>
                          {task.Status}
                        </span>

                        {task.RetryCount !== undefined && task.RetryCount > 0 && (
                          <span className="flex items-center gap-0.5 text-xs bg-orange-500 dark:bg-orange-600 text-white px-1.5 py-0.5 rounded font-bold shadow-sm">
                            <span className="animate-spin">↻</span> {task.RetryCount}/{task.MaxRetries || 3}
                          </span>
                        )}
                      </div>

                      {/* Timestamp */}
                      {task.UpdatedAt && (
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">{new Date(task.UpdatedAt).toLocaleTimeString()}</p>
                      )}
                    </div>

                    {/* Progress Bar for In progress */}
                    {task.Status === "IN_PROGRESS" && (
                      <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-gray-200 dark:bg-slate-700">
                        <div className="h-full bg-linear-to-r from-yellow-400 to-orange-400 animate-[progress_2s_ease-in-out_infinite]"></div>
                      </div>
                    )}

                    {/* Shine Effect on Hover */}
                    <div className="absolute inset-0 bg-linear-to-tr from-transparent via-white/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>
                  </div>
                </div>

                {/* Connector Arrow */}
                {index < tasks.length - 1 && (
                  <div className="flex items-center mx-4">
                    <div className="relative">
                      <div className="h-0.5 w-12 bg-linear-to-r from-gray-300 to-gray-400 dark:from-slate-600 dark:to-slate-500"></div>
                      <div className="absolute -right-1 top-1/2 -translate-y-1/2 w-0 h-0 border-t-4 border-t-transparent border-b-4 border-b-transparent border-l-8 border-l-gray-400 dark:border-l-slate-500"></div>
                    </div>
                  </div>
                )}

                {/* Next Step Indicator */}
                {index === tasks.length - 1 && workflow.Status === "IN_PROGRESS" && (
                  <div className="flex items-center mx-4 opacity-50">
                    <div className="relative">
                      <div className="h-0.5 w-12 border-t-2 border-dashed border-gray-400 dark:border-slate-600 animate-pulse"></div>
                      <div className="absolute -right-1 top-1/2 -translate-y-1/2 text-gray-400 dark:text-slate-600 animate-pulse">▶</div>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Activity Logs Section */}
        {activityLogs && activityLogs.length > 0 && (
          <div className="mt-8">
            <h2 className="text-2xl font-bold mb-4 flex items-center gap-2">
              📝 Activity Logs
              <span className="text-sm bg-purple-600 px-2 py-1 rounded-full text-white">{activityLogs.length}</span>
            </h2>
            <div className="bg-white dark:bg-slate-800 rounded-xl shadow-lg overflow-hidden transition-colors">
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-100 dark:bg-slate-700">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-semibold text-gray-700 dark:text-gray-200 uppercase tracking-wider">
                        Timestamp
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-semibold text-gray-700 dark:text-gray-200 uppercase tracking-wider">
                        Event Type
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-semibold text-gray-700 dark:text-gray-200 uppercase tracking-wider">
                        Task Name
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-semibold text-gray-700 dark:text-gray-200 uppercase tracking-wider">Details</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-slate-700">
                    {activityLogs.map((log: ActivityLog) => (
                      <tr key={log.ID} className="hover:bg-gray-50 dark:hover:bg-slate-700/50 transition-colors">
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400 font-mono">
                          {new Date(log.CreatedAt).toLocaleString()}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold ${getEventTypeBadge(log.EventType)}`}>
                            {log.EventType}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">{log.TaskName || "-"}</td>
                        <td className="px-6 py-4 text-sm text-gray-700 dark:text-gray-300">
                          <pre className="font-mono text-xs bg-gray-100 dark:bg-slate-900 p-2 rounded overflow-x-auto max-w-md">
                            {JSON.stringify(JSON.parse(log.Details || "{}"), null, 2)}
                          </pre>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Helper: เลือกสีตามสถานะ
function getStatusStyle(status: string) {
  switch (status) {
    case "COMPLETED":
      return "border-green-500 bg-white dark:bg-slate-800 shadow-lg shadow-green-500/20";
    case "IN_PROGRESS":
      return "border-yellow-400 bg-white dark:bg-slate-800 shadow-lg shadow-yellow-400/20";
    case "FAILED":
      return "border-red-500 bg-white dark:bg-slate-800 shadow-lg shadow-red-500/20";
    case "PENDING":
      return "border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 shadow-md";
    default:
      return "border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800";
  }
}

// Helper: สี header bar
function getHeaderColor(status: string) {
  switch (status) {
    case "COMPLETED":
      return "bg-gradient-to-r from-green-400 to-green-500";
    case "IN_PROGRESS":
      return "bg-gradient-to-r from-yellow-400 to-orange-400 animate-pulse";
    case "FAILED":
      return "bg-gradient-to-r from-red-400 to-red-500";
    case "PENDING":
      return "bg-gradient-to-r from-slate-300 to-slate-400";
    default:
      return "bg-slate-300";
  }
}

// Helper: พื้นหลัง icon
function getIconBg(status: string) {
  switch (status) {
    case "COMPLETED":
      return "bg-green-100 dark:bg-green-500/20";
    case "IN_PROGRESS":
      return "bg-yellow-100 dark:bg-yellow-400/20";
    case "FAILED":
      return "bg-red-100 dark:bg-red-500/20";
    case "PENDING":
      return "bg-slate-100 dark:bg-slate-700";
    default:
      return "bg-slate-100 dark:bg-slate-700";
  }
}

// Helper: status badge
function getStatusBadge(status: string) {
  switch (status) {
    case "COMPLETED":
      return "bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-400";
    case "IN_PROGRESS":
      return "bg-yellow-100 text-yellow-700 dark:bg-yellow-400/20 dark:text-yellow-400";
    case "FAILED":
      return "bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-400";
    case "PENDING":
      return "bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400";
    default:
      return "bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400";
  }
}

// Helper: status dot
function getStatusDot(status: string) {
  switch (status) {
    case "COMPLETED":
      return "bg-green-500";
    case "IN_PROGRESS":
      return "bg-yellow-500";
    case "FAILED":
      return "bg-red-500";
    case "PENDING":
      return "bg-slate-400";
    default:
      return "bg-slate-400";
  }
}

// Helper: เลือกไอคอน
function getIcon(status: string, size: number = 28) {
  switch (status) {
    case "COMPLETED":
      return <CheckCircle size={size} className="text-green-500 dark:text-green-400" />;
    case "IN_PROGRESS":
      return <Clock size={size} className="animate-spin text-yellow-500 dark:text-yellow-400" />;
    case "FAILED":
      return <AlertTriangle size={size} className="text-red-500 dark:text-red-400" />;
    default:
      return <Circle size={size} className="text-slate-400 dark:text-slate-500" />;
  }
}

// Helper: event type badge color
function getEventTypeBadge(eventType: string) {
  switch (eventType) {
    case "TASK_STARTED":
      return "bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-400";
    case "TASK_COMPLETED":
      return "bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-400";
    case "WORKFLOW_COMPLETED":
      return "bg-purple-100 text-purple-700 dark:bg-purple-500/20 dark:text-purple-400";
    case "TASK_FAILED":
      return "bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-400";
    default:
      return "bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400";
  }
}
