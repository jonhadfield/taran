"use client";

import { Component, type ReactNode } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { AlertCircle } from "lucide-react";

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
}

/**
 * Error boundary for individual dashboard widgets. If a chart or card
 * throws during render, only that widget shows an error — the rest of
 * the dashboard stays functional.
 */
export class WidgetErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? (
        <Card>
          <CardContent className="flex items-center gap-2 py-4 text-sm text-muted-foreground">
            <AlertCircle className="size-4 shrink-0" />
            <span>Failed to load this widget.</span>
          </CardContent>
        </Card>
      );
    }
    return this.props.children;
  }
}
