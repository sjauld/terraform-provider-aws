package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/service/medialive"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	audioSelectorTypes = []string{"language", "pid", "track"}
	captionSourceTypes = []string{"arib", "dvb_source", "embedded", "scte20", "scte27", "teletext"}
	outputTypes        = []string{"media_package", "multiplex", "standard"}
	outputGroupTypes   = []string{
		"archive", "frame_capture", "hls", "media_package", "ms_smooth",
		"multiplex", "rtmp", "udp"}
)

func resourceAwsMediaLiveChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaLiveChannelCreate,
		Read:   resourceAwsMediaLiveChannelRead,
		Update: resourceAwsMediaLiveChannelUpdate,
		Delete: resourceAwsMediaLiveChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"class": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(medialive.ChannelClass_Values(), false),
			},
			"destination": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"type": {
							Required:     true,
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice(outputTypes, false),
						},
						// MediaPackage
						"media_package_channel_ids": {
							Optional: true,
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						// Multiplex
						"multiplex_id": {
							Optional: true,
							Type:     schema.TypeString,
						},
						// Multiplex
						"multiplex_program_name": {
							Optional: true,
							Type:     schema.TypeString,
						},
						// Standard
						"password_param": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"stream_name": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"url": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"username": {
							Optional: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			// "encoder_settings": {
			// 	Type:     schema.TypeList,
			// 	MaxItems: 1,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			// only the required settings are supported at the moment
			// 			// "audio_descriptions":  {},
			// 			// "avail_blanking":      {},
			// 			// "avail_configuration": {},
			// 			// "blackout_state":      {},
			// 			// "caption_descriptions": {},
			// 			// "feature_activations": {},
			// 			// "global_configuration": {},
			// 			// "nielsen_configuration": {},
			// 			"output_groups": {
			// 				Required: true,
			// 				Type:     schema.TypeSet,
			// 				Elem: &schema.Resource{
			// 					Schema: map[string]*schema.Schema{
			// 						"name": {
			// 							Optional: true,
			// 							Type:     schema.TypeString,
			// 						},
			// 						"type": {
			// 							Required:     true,
			// 							Type:         schema.TypeString,
			// 							ValidateFunc: validation.StringInSlice(outputGroupTypes, false),
			// 						},
			// 						// // archive
			// 						// "rollover_interval": {
			// 						// 	Optional: true,
			// 						// 	Type:     schema.TypeInt,
			// 						// },
			// 						// archive, frame_capture, media_package,
			// 						"destination_ref_id": {
			// 							Required: true,
			// 							Type:     schema.TypeString,
			// 						},
			// 						// hls - not supported yet
			// 						// "hls_group_settings": {}
			// 						// ms smooth - not supported yet
			// 						// "ms_smooth_settings": {}
			// 						// rtmp - not supported yet
			// 						// "rtmp_group_settings": {}
			// 						// udp - not supported yet
			// 						// "udp_group_settings": {}
			// 						"outputs": {
			// 							// @TODO this is where I am up to
			// 							Required: true,
			// 							Type:     schema.TypeSet,
			// 							Elem: &schema.Resource{
			// 								Schema: map[string]*schema.Schema{
			// 									"audio_description_names": {
			// 										Optional: true,
			// 										Type:     schema.TypeSet,
			// 										Elem:     &schema.Schema{Type: schema.TypeString},
			// 									},
			// 									"caption_description_names": {
			// 										Optional: true,
			// 										Type:     schema.TypeSet,
			// 										Elem:     &schema.Schema{Type: schema.TypeString},
			// 									},
			// 									"name": {
			// 										Optional: true,
			// 										Type:     schema.TypeString,
			// 									},
			// 									"video_description_name": {
			// 										Optional: true,
			// 										Type:     schema.TypeString,
			// 									},
			// 									// // archive - not supported yet
			// 									// // "m2ts_settings_settings": {},
			// 									// // archive
			// 									// "extension": {
			// 									// 	Optional: true,
			// 									// 	Type:     schema.TypeString,
			// 									// },
			// 									// // archive, frame capture, hls, MS smooth
			// 									// "name_modifier": {
			// 									// 	Optional: true,
			// 									// 	Type:     schema.TypeString,
			// 									// },
			// 									// // hls, MS smooth
			// 									// "h265_packaging_type": {
			// 									// 	Optional:     true,
			// 									// 	Type:         schema.TypeString,
			// 									// 	ValidateFunc: validation.StringInSlice(medialive.HlsH265PackagingType_Values(), false),
			// 									// },
			// 									// // hls
			// 									// "segment_modifier": {
			// 									// 	Optional: true,
			// 									// 	Type:     schema.TypeString,
			// 									// },
			// 									// // hls
			// 									// "audio_rendition_sets": {
			// 									// 	Optional: true,
			// 									// 	Type:     schema.TypeString,
			// 									// },
			// 									// // hls - not supported yet
			// 									// // "m3u8_settings": {}
			// 									// // multiplex, rtmp
			// 									// "destination_ref_id": {
			// 									// 	Optional: true,
			// 									// 	Type:     schema.TypeString,
			// 									// },
			// 									// // rtmp - other fields not supported yet
			// 								},
			// 							},
			// 						},
			// 					},
			// 				},
			// 			},
			// 			"timecode_config": {
			// 				Required: true,
			// 			},
			// 			"video_descriptions": {
			// 				Required: true,
			// 			},
			// 		},
			// 	},
			// },
			"input_attachment": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automatic_input_failover_settings": {
							Optional: true,
							Type:     schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input_preference": {
										Required:     true,
										Type:         schema.TypeString,
										ValidateFunc: validation.StringInSlice(medialive.InputPreference_Values(), false),
									},
									"secondary_input_id": {
										Required: true,
										Type:     schema.TypeString,
									},
								},
							},
						},
						"name": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"input_id": {
							Required: true,
							Type:     schema.TypeString,
						},
						"input_settings": {
							Optional: true,
							Type:     schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"audio_selector": {
										Optional: true,
										Type:     schema.TypeSet,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Required: true,
													Type:     schema.TypeString,
												},
												"type": {
													Required:     true,
													Type:         schema.TypeString,
													ValidateFunc: validation.StringInSlice(audioSelectorTypes, false),
												},
												// language
												"language_code": {
													Required: true,
													Type:     schema.TypeString,
												},
												// language
												"language_selection_policy": {
													Required:     true,
													Type:         schema.TypeString,
													ValidateFunc: validation.StringInSlice(medialive.AudioLanguageSelectionPolicy_Values(), false),
												},
												// pid
												"pid": {
													Optional: true,
													Type:     schema.TypeInt,
												},
												// track
												"tracks": {
													Required: true,
													Type:     schema.TypeList,
													Elem:     &schema.Schema{Type: schema.TypeInt},
												},
											},
										},
									},
									"caption_selector": {
										Optional: true,
										Type:     schema.TypeSet,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Required: true,
													Type:     schema.TypeString,
												},
												"language_code": {
													Required: true,
													Type:     schema.TypeString,
												},
												"type": {
													Required:     true,
													Type:         schema.TypeString,
													ValidateFunc: validation.StringInSlice(captionSourceTypes, false),
												},
												// DVB Sub / SCTE27
												"pid": {
													Optional: true,
													Type:     schema.TypeString,
												},
												// embedded / SCTE20
												"upconvert_608": {
													Optional: true,
													Default:  false,
													Type:     schema.TypeBool,
												},
												// embedded
												"detect_scte20": {
													Optional: true,
													Default:  false,
													Type:     schema.TypeBool,
												},
												// embedded / SCTE20
												"channel": {
													Optional: true,
													Type:     schema.TypeInt,
												},
												// teletext
												"page": {
													Optional: true,
													Type:     schema.TypeString,
												},
											},
										},
									},
									"deblock_filter": {
										Optional:     true,
										Type:         schema.TypeString,
										ValidateFunc: validation.StringInSlice(medialive.InputDeblockFilter_Values(), false),
									},
									"denoise_filter": {
										Optional:     true,
										Type:         schema.TypeString,
										ValidateFunc: validation.StringInSlice(medialive.InputDenoiseFilter_Values(), false),
									},
									"filter_strength": {
										Optional:     true,
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntBetween(1, 5),
									},
									"input_filter": {
										Optional:     true,
										Type:         schema.TypeString,
										ValidateFunc: validation.StringInSlice(medialive.InputFilter_Values(), false),
									},
									"network_input_setting": {
										Optional: true,
										Type:     schema.TypeSet,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"server_validation": {
													Optional:     true, // maybe? the docs are awful
													Type:         schema.TypeString,
													ValidateFunc: validation.StringInSlice(medialive.NetworkInputServerValidation_Values(), false),
												},
												"hls_bandwidth": {
													Optional: true,
													Type:     schema.TypeInt,
												},
												"hls_buffer_segments": {
													Optional: true,
													Type:     schema.TypeInt,
												},
												"hls_retries": {
													Optional: true,
													Type:     schema.TypeInt,
												},
												"hls_retry_interval": {
													Optional: true,
													Type:     schema.TypeInt,
												},
											},
										},
									},
									"smpte_2038_data_preference": {
										Optional:     true,
										Type:         schema.TypeString,
										ValidateFunc: validation.StringInSlice(medialive.Smpte2038DataPreference_Values(), false),
									},
									"source_end_behaviour": {
										Optional:     true,
										Type:         schema.TypeString,
										ValidateFunc: validation.StringInSlice(medialive.InputSourceEndBehavior_Values(), false),
									},
									"video_selector": {
										Optional: true,
										Type:     schema.TypeSet,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"color_space": {
													Required:     true,
													Type:         schema.TypeString,
													ValidateFunc: validation.StringInSlice(medialive.VideoSelectorColorSpace_Values(), false),
												},
												"color_space_usage": {
													Required:     true,
													Type:         schema.TypeString,
													ValidateFunc: validation.StringInSlice(medialive.VideoSelectorColorSpaceUsage_Values(), false),
												},
												"pid": {
													Optional: true,
													Type:     schema.TypeInt,
												},
												"program_id": {
													Optional: true,
													Type:     schema.TypeInt,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"log_level": {
				Optional:     true,
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(medialive.LogLevel_Values(), false),
			},
			"name": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"role_arn": {
				Optional: true,
				Type:     schema.TypeString,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsMediaLiveChannelCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] creating channel")
	return resourceAwsMediaLiveChannelRead(d, meta)
}
func resourceAwsMediaLiveChannelRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}
func resourceAwsMediaLiveChannelUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsMediaLiveChannelRead(d, meta)
}
func resourceAwsMediaLiveChannelDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
