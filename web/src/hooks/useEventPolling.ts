import { useCallback, useEffect, useRef, useState } from 'react';
import { api, type Event } from '../api/client';
import { isEventActive } from '../lib/format';

const ACTIVE_POLL_MS = 1500;

export function useEventPolling(eventId: string | undefined) {
  const [event, setEvent] = useState<Event | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [, setTick] = useState(0);

  const load = useCallback(async () => {
    if (!eventId) return;
    const ev = await api.getEvent(eventId);
    setEvent(ev);
    setLastUpdated(new Date());
    return ev;
  }, [eventId]);

  const refresh = useCallback(async () => {
    try {
      await load();
      setError('');
    } catch (e) {
      setError(String(e));
    }
  }, [load]);

  useEffect(() => {
    setLoading(true);
    load()
      .catch((e) => setError(String(e)))
      .finally(() => setLoading(false));
  }, [load]);

  useEffect(() => {
    if (!event || !isEventActive(event)) return;
    const timer = window.setInterval(() => {
      load().catch(() => {});
    }, ACTIVE_POLL_MS);
    return () => window.clearInterval(timer);
  }, [event?.status, load]);

  useEffect(() => {
    if (!lastUpdated || !event || !isEventActive(event)) return;
    const timer = window.setInterval(() => setTick((n) => n + 1), 5000);
    return () => window.clearInterval(timer);
  }, [lastUpdated, event?.status]);

  return { event, setEvent, error, setError, loading, lastUpdated, refresh, load };
}

export function useAutoSelectJob(jobs: Event['jobs'], selectedId: string | null, setSelectedId: (id: string) => void) {
  const prevSig = useRef('');

  useEffect(() => {
    if (!jobs?.length) return;
    const sig = jobs.map((j) => `${j.id}:${j.status}`).join('|');
    if (sig === prevSig.current && selectedId && jobs.some((j) => j.id === selectedId)) return;
    prevSig.current = sig;

    const active = jobs.find((j) => j.status === 'RUNNING' || j.status === 'QUEUED');
    const failed = jobs.find((j) => j.status === 'FAILED');
    const pick = active ?? failed ?? jobs[0];
    if (pick && pick.id !== selectedId) setSelectedId(pick.id);
  }, [jobs, selectedId, setSelectedId]);
}
