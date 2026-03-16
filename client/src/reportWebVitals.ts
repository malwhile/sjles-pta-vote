type ReportCallback = (metric: any) => void;

const reportWebVitals = (onPerfEntry?: ReportCallback): void => {
  if (onPerfEntry && onPerfEntry instanceof Function) {
    import('web-vitals').then((webVitals) => {
      if ('getCLS' in webVitals) (webVitals as any).getCLS(onPerfEntry);
      if ('getFID' in webVitals) (webVitals as any).getFID(onPerfEntry);
      if ('getFCP' in webVitals) (webVitals as any).getFCP(onPerfEntry);
      if ('getLCP' in webVitals) (webVitals as any).getLCP(onPerfEntry);
      if ('getTTFB' in webVitals) (webVitals as any).getTTFB(onPerfEntry);
    });
  }
};

export default reportWebVitals;
