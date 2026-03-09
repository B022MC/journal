import { ReactRoadmap } from "@/components/roadmap";

export default function Home() {
  return (
    <div className="min-h-screen bg-black text-neutral-200 selection:bg-neutral-800 selection:text-white">
      <main>
        <ReactRoadmap />
      </main>
    </div>
  );
}
