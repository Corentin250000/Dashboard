type AuthCardProps = {
  title: string;
  children: React.ReactNode;
};

export default function AuthCard({ title, children }: AuthCardProps) {
  return (
    <div className="flex min-h-screen items-center justify-center px-6">
      <div
        className="
          w-full
          max-w-sm
          rounded-xl
          bg-(--background)
          border
          border-(--border)
          p-8
        "
      >
        <h1 className="mb-7 text-[22px] font-medium text-(--text-secondary)">
          {title}
        </h1>

        {children}
      </div>
    </div>
  );
}