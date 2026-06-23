import { useTranslation } from 'react-i18next'
import { PieChart } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  ChartLegend,
  ChartLegendContent,
} from '@/components/ui/chart'
import { PieChart as RePieChart, Pie, Cell } from 'recharts'
import { EmptyState } from '@/components/empty-state'

interface RiskDistributionChartProps {
  data: {
    low: number
    medium: number
    high: number
    critical: number
  }
}

export function RiskDistributionChart({ data }: RiskDistributionChartProps) {
  const { t } = useTranslation()

  const chartData = [
    { key: 'low', label: t('Low'), value: data.low, color: 'hsl(var(--chart-1))' },
    { key: 'medium', label: t('Medium'), value: data.medium, color: 'hsl(var(--chart-2))' },
    { key: 'high', label: t('High'), value: data.high, color: 'hsl(var(--chart-3))' },
    { key: 'critical', label: t('Critical'), value: data.critical, color: 'hsl(var(--chart-4))' },
  ].filter((d) => d.value > 0)

  const config = {
    low: { label: t('Low'), color: 'hsl(var(--chart-1))' },
    medium: { label: t('Medium'), color: 'hsl(var(--chart-2))' },
    high: { label: t('High'), color: 'hsl(var(--chart-3))' },
    critical: { label: t('Critical'), color: 'hsl(var(--chart-4))' },
  }

  const total = data.low + data.medium + data.high + data.critical

  return (
    <Card className="flex flex-col">
      <CardHeader className="flex flex-row items-center gap-2">
        <PieChart className="size-4 text-muted-foreground" />
        <CardTitle>{t('Risk Distribution')}</CardTitle>
      </CardHeader>
      <CardContent className="flex-1">
        {total === 0 ? (
          <EmptyState
            icon={PieChart}
            title={t('No Data')}
            description={t('No risk distribution data for the selected period.')}
            className="min-h-[200px] rounded-lg border border-dashed"
            bordered={false}
          />
        ) : (
          <ChartContainer config={config} className="min-h-[260px]">
            <RePieChart>
              <Pie
                data={chartData}
                dataKey="value"
                nameKey="key"
                cx="50%"
                cy="50%"
                innerRadius={70}
                outerRadius={90}
                paddingAngle={2}
              >
                {chartData.map((entry) => (
                  <Cell key={entry.key} fill={entry.color} />
                ))}
              </Pie>
              <ChartTooltip
                content={
                  <ChartTooltipContent
                    formatter={(value: any, name: any) => (
                      <span>
                        {config[name as keyof typeof config]?.label ?? name}: {value}
                      </span>
                    )}
                  />
                }
              />
              <ChartLegend content={<ChartLegendContent />} />
            </RePieChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  )
}
