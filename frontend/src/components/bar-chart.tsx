interface BarChartProps {
  data: { label: string; value: number }[];
  height?: number;
  formatValue?: (n: number) => string;
}

export function BarChart({ data, height = 80, formatValue }: BarChartProps) {
  const max = Math.max(...data.map((d) => d.value), 1);
  const fmt = formatValue || ((n: number) => String(n));
  return (
    <div
      className="grid gap-1"
      style={{ gridTemplateColumns: `repeat(${data.length}, 1fr)`, height }}
    >
      {data.map((d, i) => {
        const pct = Math.max(4, (d.value / max) * 100);
        return (
          <div key={i} className="flex flex-col justify-end">
            <div
              className="w-full rounded-sm bg-primary/70 hover:bg-primary transition-colors"
              style={{ height: `${pct}%` }}
              title={`${d.label}: ${fmt(d.value)}`}
            />
          </div>
        );
      })}
    </div>
  );
}
