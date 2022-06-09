package core

//
// func (s *Service) syncChange(ctx context.Context, res resource.Resource) (*resource.Resource, bool, error) {
// 	modSpec, err := s.generateModuleSpec(ctx, res)
// 	if err != nil {
// 		return nil, false, err
// 	}
//
// 	oldState := res.State.Clone()
// 	newState, err := s.rootModule.Sync(ctx, *modSpec)
// 	if err != nil {
// 		if errors.Is(err, errors.ErrInvalid) {
// 			return nil, false, err
// 		}
// 		return nil, false, errors.ErrInternal.WithMsgf("sync() failed").WithCausef(err.Error())
// 	}
//
// 	// TODO: clarify on behaviour when resource schedule for deletion reaches error.
// 	shouldDelete := oldState.InDeletion() && newState.IsTerminal()
//
// 	res.UpdatedAt = s.clock()
// 	res.State = *newState
// 	return &res, shouldDelete, nil
// }
