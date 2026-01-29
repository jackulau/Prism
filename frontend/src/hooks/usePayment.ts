import { trpc } from '../lib/trpc';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const t = trpc as any;

// Plans
export const usePlans = () => {
  return t.payment.listPlans.useQuery();
};

export const useCurrentPlan = () => {
  return t.payment.getCurrentPlan.useQuery();
};

// Subscriptions
export const useSubscription = () => {
  return t.payment.getSubscription.useQuery();
};

export const useCreateCheckout = () => {
  return t.payment.createCheckoutSession.useMutation({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onSuccess: (data: any) => {
      if (data.url) {
        window.location.href = data.url;
      }
    },
  });
};

export const useCancelSubscription = () => {
  const utils = t.useUtils();
  return t.payment.cancelSubscription.useMutation({
    onSuccess: () => {
      utils.payment.getSubscription.invalidate();
      utils.payment.getCurrentPlan.invalidate();
    },
  });
};

export const useReactivateSubscription = () => {
  const utils = t.useUtils();
  return t.payment.reactivateSubscription.useMutation({
    onSuccess: () => {
      utils.payment.getSubscription.invalidate();
      utils.payment.getCurrentPlan.invalidate();
    },
  });
};

// Billing Portal
export const useCreatePortalSession = () => {
  return t.payment.createPortalSession.useMutation({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onSuccess: (data: any) => {
      if (data.url) {
        window.location.href = data.url;
      }
    },
  });
};

// Usage
export const useUsage = () => {
  return t.payment.getUsage.useQuery();
};

// Invoices
export const useInvoices = () => {
  return t.payment.listInvoices.useQuery();
};

// Payment Methods
export const usePaymentMethods = () => {
  return t.payment.listPaymentMethods.useQuery();
};

export const useSetDefaultPaymentMethod = () => {
  const utils = t.useUtils();
  return t.payment.setDefaultPaymentMethod.useMutation({
    onSuccess: () => {
      utils.payment.listPaymentMethods.invalidate();
      utils.payment.getSubscription.invalidate();
    },
  });
};

export const useDeletePaymentMethod = () => {
  const utils = t.useUtils();
  return t.payment.deletePaymentMethod.useMutation({
    onSuccess: () => {
      utils.payment.listPaymentMethods.invalidate();
    },
  });
};
