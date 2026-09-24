interface BarChartProps {
  data: { label: string; value: number }[];
  height?: number;
  formatValue?: (n: number) => string;
}

export function BarChart({ data, height = 80, formatValue }: BarChartProps) {
  const max = Math.max(...data.map((d) => d.value), 1);
  const fmt = formatValue || ((n: number) => String(n));
  return (
    // A baseline gives the bars something to stand on, so a single week reads
    // as one column rather than a block of colour filling the panel.
    <div
      className="flex items-end justify-center gap-0.5 border-b"
      style={{ height }}
    >
      {data.map((d, i) => {
        const pct = Math.max(4, (d.value / max) * 100);
        return (
          <div key={i} className="flex h-full max-w-14 flex-1 flex-col justify-end">
            <div
              className="w-full rounded-t-[4px] bg-gradient-to-t from-primary to-primary/65 transition-opacity hover:opacity-90"
              style={{ height: `${pct}%` }}
              title={`${d.label}: ${fmt(d.value)}`}
            />
          </div>
        );
      })}
    </div>
  );
}
