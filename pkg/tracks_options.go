package mediasource

type TracksOption = func(*Tracks) error

// NOTE: THIS SHOULD BE USED WITH WithTWCCSenderInterceptor and the interval needs to be same

// func WithBandwidthEstimatorInterceptor(initialBitrate int, interval time.Duration) TracksOption {
// 	return func(tracks *Tracks) error {
// 		var (
// 			congestionController *cc.InterceptorFactory
// 			err                  error
// 		)
//
// 		tracks.bwEstimator = &bandwidthEstimator{ctx: tracks.ctx, consumers: make(map[string]*consumer), interval: interval}
//
// 		if congestionController, err = cc.NewInterceptor(func() (cc.BandwidthEstimator, error) {
// 			return gcc.NewSendSideBWE(gcc.SendSideBWEInitialBitrate(initialBitrate))
// 		}); err != nil {
// 			return err
// 		}
//
// 		congestionController.OnNewPeerConnection(func(id string, estimator cc.BandwidthEstimator) {
// 			fmt.Printf("got bitrate estimator for peer connection with label: %s\n", id)
// 			tracks.bwEstimator.estimator = estimator
// 			tracks.bwEstimator.Start()
// 		})
//
// 		tracks.interceptorRegistry.Add(congestionController)
// 		if err = webrtc.ConfigureTWCCHeaderExtensionSender(tracks.mediaEngine, tracks.interceptorRegistry); err != nil {
// 			return err
// 		}
//
// 		return nil
// 	}
// }
