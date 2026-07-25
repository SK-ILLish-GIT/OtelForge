type Props = {
  verified: number;
  failed: number;
  total: number;
};

export function ProgressBar({ verified, failed, total }: Props) {
  if (total <= 0) return null;
  const verifiedPct = (verified / total) * 100;
  const failedPct = (failed / total) * 100;
  return (
    <div className="progress-bar" role="progressbar" aria-valuenow={verified + failed} aria-valuemin={0} aria-valuemax={total}>
      <div className="progress-segment progress-verified" style={{ width: `${verifiedPct}%` }} />
      <div className="progress-segment progress-failed" style={{ width: `${failedPct}%` }} />
    </div>
  );
}
