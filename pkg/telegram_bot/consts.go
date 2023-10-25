package telegram_bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	startText             = "Welcome to our identity verification chatbot, powered by DataSpike.io! \nWe understand the importance of keeping your personal data safe and secure, which is why we want to assure you that we do not cache any data. \nYour information will be automatically removed from the chat within 1 hour for your privacy and security. \nIf you have any questions or concerns about the verification process, please don't hesitate to contact us. \n\nWe're here to help.\n\n/start_verification - Start new document verification process\n/help - Display help information\n/cancel - Cancel ongoing verification\n/ask_expert - Ask AI expert\n/customize_bot - Integrate bot to your platform"
	helpText              = "Thank you for using our KYC verification bot! To ensure a smooth and easy verification process, please read the following instructions carefully:\n\n- What is MRZ? The Machine Readable Zone (MRZ) is a series of characters found on most passports and IDs that contain important personal information. - Please ensure that your ID document contains an MRZ before uploading it.\nWhich documents support MRZ? Most passports and government-issued IDs, such as driver's licenses, national ID cards, and residence permits, contain an MRZ. Please check your document to confirm.\n- How to upload high-quality photos? For the best results, please ensure that your photos are clear, in focus, and well-lit. Avoid shadows and glare, and make sure all text and information is visible and legible.\n\nIf you encounter any issues during the verification process or have any questions, please don't hesitate to contact us for assistance.\n\n<a href='https://www.dataspike.io/contact-us'>Contact us</a>"
	cancelText            = "Your verification process has been cancelled. If you need to verify your identity in the future, please don't hesitate to start the process again. If you encountered any issues or have any questions, please feel free to contact us for assistance.\n<a href='https://www.dataspike.io/contact-us'>Contact us</a>"
	askExpertText         = "Hello and welcome!\n\nAs an AI expert in KYC, I am here to assist you with any questions you may have and help guide you through the KYC process. Whether you're new to KYC or a seasoned professional, I am here to provide you with the expertise and support you need to successfully complete your KYC requirements.\n\nPlease don't hesitate to ask me any questions you may have. I am always here to help and ensure your KYC experience is as smooth and hassle-free as possible."
	customizeBotText      = "Are you looking to integrate identity verification into your chatbot, but not sure where to start?\nAt Dataspike, we can help! Our KYC solution is designed to be easily integrated into any messaging platform, including Telegram, WhatsApp, and Facebook Messenger. \nAnd if you need a custom verification bot built to fit your unique requirements, we're happy to help with that too. \nWhether you're looking to streamline your customer onboarding process or improve security for your users, \nDataspike has the tools and expertise you need to get the job done. <a href='https://www.dataspike.io/contact-us'>Contact us</a> today to learn more!"
	startVerificationInit = "To get started, we need to verify your identity. Please follow the prompts below to complete the process."
	poiAttachDocument     = "Please attach photo of your document so we can verify its authenticity.\n\nYou can attach a Passport, residence permit, or national ID card. Provided document must have MRZ code."
	poiHelpForButton      = "Please note that we currently only accept documents with an MRZ code for our verification process."
	mzrText               = "MRZ (Machine Readable Zone) is a specific type of code found on certain identity documents, such as passports and visas.\nIt contains important information about the document holder, including their name, nationality, and date of birth.\nIf you`re not sure if your document has an MRZ, look for two lines of text at the bottom of the personal information page.\nThe first line will have letters and the second line will have numbers. This is the MRZ code."
	expiredText           = "Your verification is expired.\nPlease contact us for assistance to make new verification.\n\n<a href='https://www.dataspike.io/contact-us'>Contact us</a>"
	verificationNotFound  = `Your verification is not found. 
Please contact us for assistance to make new verification.

<a href='https://www.dataspike.io/contact-us'>Contact us</a>

You can choose from a utility bill, residence registration or bank statement.`
	PoaSkipPrompt      = "To skip this step press skip button."
	AttachSelfiePrompt = `To verify your identity, we need to see a photo of you.
Please attach a selfie so we can confirm your identity.`
	LivenessPrompt = `To verify your identity, we need to see a photo of you.
Please click button below to open liveness widget.`
	VerificationStartedPleaseWait = `We are now processing your verification. This may take a few minutes.
You will receive a notification as soon as the process is complete.`
	AttachBackSideOfDoc = "Please attach <b>the back side</b> of your document so we can verify its authenticity."
	SelectedPoaDocument = "Please attach poa document."
	VerificationOk      = "Congratulations! Your identity has been verified."
	VerificationFailed  = `We're sorry, but your verification has failed. Please contact us for assistance to make new verification.

<a href='https://www.dataspike.io/contact-us'>Contact us</a>`
	LivenessFailed               = `We're sorry, but step with liveness photo has failed. Please check the reason below and try again.`
	verificationForBotIsDisabled = "Verification for bots is disabled."
	verificationCompleted        = "Your verification is completed."
)

const (
	Selfie   = "selfie"
	Poi      = "poi"
	Poa      = "poa"
	Front    = "front"
	learnMrz = "learn_more_about_mzr"
	skipPoa  = "skip_poa"
	pending  = "pending"
	verified = "verified"

	mrzLink = "https://static.dataspike.io/images/docver/mrz_sample.jpg"

	buffer = 100
)

var defaultProfile = "1ee20e99-35f7-6c75-811b-6df0f88c424d"

var contactUsKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Contact Dataspike", "https://www.dataspike.io/contact-us"),
	),
)

var menu = tgbotapi.NewSetMyCommands(
	tgbotapi.BotCommand{
		Command:     "/help",
		Description: "Help",
	},
	tgbotapi.BotCommand{
		Command:     "/start_verification",
		Description: "Start verification",
	},
	tgbotapi.BotCommand{
		Command:     "/customize_bot",
		Description: "Customize bot",
	},
	tgbotapi.BotCommand{
		Command:     "/ask_expert",
		Description: "Ask expert",
	},
	tgbotapi.BotCommand{
		Command:     "/cancel",
		Description: "Cancel",
	},
)

func generateLivenessKeyboard(url string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Liveness", url),
		),
	)
}

var mzrKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Learn more about MZR", learnMrz),
	),
)

var skipPoaKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("skip", skipPoa),
	),
)
