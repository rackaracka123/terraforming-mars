interface FormattedDescriptionProps {
  text: string;
}

export const FormattedDescription = ({ text }: FormattedDescriptionProps) => {
  const parts = text.split(/\*\*(.*?)\*\*/);
  return <>{parts.map((part, i) => (i % 2 === 1 ? <strong key={i}>{part}</strong> : part))}</>;
};
