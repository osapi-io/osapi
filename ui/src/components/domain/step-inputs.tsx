import { StepForm } from "@/components/domain/step-form";
import { ObjectUploadForm } from "@/components/domain/object-upload-form";
import { ObjectDeleteForm } from "@/components/domain/object-delete-form";
import type { StepDef } from "@/lib/catalog-index";
import type { PlanStep } from "@/lib/plan";

interface StepInputsProps {
  step: PlanStep;
  def: StepDef;
  onChange: (data: Record<string, unknown>) => void;
}

/**
 * The fields of a step: the form of a behavior's method, made from its
 * input, or the form of an Object Store step.
 */
export function StepInputs({ step, def, onChange }: StepInputsProps) {
  if (def.method) {
    return (
      <StepForm
        stepId={step.id}
        method={def.method}
        data={step.with}
        onChange={onChange}
      />
    );
  }
  if (step.run === "objectstore.upload") {
    return <ObjectUploadForm data={step.with} onChange={onChange} />;
  }
  if (step.run === "objectstore.delete") {
    return (
      <ObjectDeleteForm stepId={step.id} data={step.with} onChange={onChange} />
    );
  }
  return null;
}
