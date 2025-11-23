"use client";

import Link from "next/link";
import ThemeToggle from "./ThemeToggle";
import { Workflow } from "lucide-react";

export default function Navbar() {
  return (
    <nav className="sticky top-0 z-50 bg-white dark:bg-slate-900 border-b border-gray-200 dark:border-slate-700 shadow-sm transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo */}
          <Link href="/" className="flex items-center gap-2 group">
            <Workflow className="w-8 h-8 text-blue-600 dark:text-blue-400 group-hover:rotate-12 transition-transform" />
            <span className="text-xl font-bold text-gray-900 dark:text-white">Go-Flow</span>
          </Link>

          {/* Nav Items */}
          <div className="flex items-center gap-6">
            <Link
              href="/"
              className="text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
              Workflows
            </Link>
            <ThemeToggle />
          </div>
        </div>
      </div>
    </nav>
  );
}
