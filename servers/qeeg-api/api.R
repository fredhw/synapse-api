library(e1071)
library(pracma)
library(jsonlite)
library(openxlsx)
library(base64enc)
source('eeg.analysis.3.1.3.R')

#* @filter cors
cors <- function(res) {
    res$setHeader("Access-Control-Allow-Origin", "*")
    plumber::forward()
}

#' Echo the parameter that was sent in
#' @param msg The message to echo back.
#' @preempt cors
#' @get /echo
function(msg=""){
  list(msg = paste0("The message is: '", msg, "'"))
}

#' @get /v1/hello
#' @preempt cors
#' @html
function(){
  "<html><h1>hello world</h1></html>"
}

#' Create a summary file
#' @param subject The message to echo back.
#' @param session The session name
#' @preempt cors
#' @serializer unboxedJSON
#' @get /v1/sumfile/
function(req, subject, session, sampling=128, window=2, sliding=0.75) {	
	channels <- c("AF3", "F7", "F3", "FC5", 
                "T7", "P7", "O1", "O2", 
                "P8", "T8", "FC6", "F4", 
                "F8", "AF4")
	print("obtaining user data from header:")
	print(paste0("xuser=",req$HTTP_X_USER))
	json <- fromJSON(req$HTTP_X_USER)
  	print(paste0("userName: ", json$userName))



	file <- paste("./raw-data/", json$userName, "/" ,subject, "_", session, ".txt", sep="")
  
	if ( file.exists(file) ) {
		data <- read.table(file, header=T)
		samples <- dim(data)[1]
		result <- list("Subject"=subject, "Version" = version, "Session"=session, "Sampling"=sampling,
						"Window"=window, "Sliding"=sliding, "Duration" = (samples / sampling), "Blinks" = "NA")
		
		if ("Blink" %in% names(data)) {
			blink <- data$Blink
			blink_onsets <- blink[2 : samples] - blink[1 : (samples - 1)]
			result["Meta_Blinks"] <- sum(blink_onsets[blink_onsets > 0])
		} else {
			blink <- rep(0, samples)
		}
		x <- data$GyroX[1 : samples]
		y <- data$GyroY[1 : samples]
		
		for (ch in channels) {
			#print(ch)
			ts <- data[[ch]]
			ts <- ts[1 : samples]
			qty <- data[[paste(ch, "Q", sep="_")]]
			qty <- qty[1 : samples]
			spectrum <- spectral.analysis(ts, sampling, length=window, sliding=0.75, hamming=T,
											x=x, y=y, blink=blink, quality=qty)
			
			for (j in 1:length(band.names)) {
				result[paste(ch, "_mean_", band.names[j], "_power", sep="")] <- mean.power(spectrum, bands[j,])
			}
			result[paste(ch, "IAF", sep="_")] <- iaf(spectrum)
			result[paste(ch, "IAF", "Power", sep="_")] <- iaf.power(spectrum)
			result[paste("Meta", ch, "Samples", sep="_")] <- spectrum$Samples
			result[paste("Meta", ch, "LongestQualitySegment", sep="_")] <- spectrum$LongestQualitySegment
			result[paste("Meta", ch, "SpectralQuality", sep="_")] <- spectral.quality(spectrum)
		}
		
		# ## Coherence analysis
		# for (i in  1 : (length(channels) - 1)) {
		# 	for (j in  (i + 1) : length(channels)) {
		# 		ch1 <- channels[i]
		# 		ch2 <- channels[j]
				
		# 		ts1 <- data[[ch1]]
		# 		ts2 <- data[[ch2]]
				
		# 		ts1 <- ts1[1 : samples]
		# 		ts2 <- ts2[1 : samples]
				
		# 		qty1 <- data[[paste(ch1, "Q", sep="_")]]
		# 		qty1 <- qty1[1 : samples]
				
		# 		qty2 <- data[[paste(ch2, "Q", sep="_")]]
		# 		qty2 <- qty2[1 : samples]
		# 		#print(paste("Coherence", ch1, ch2))
		# 		cohr <- coherence.analysis(ts1, ts2, sampling, length=window, sliding=0.75, hamming=T,
		# 										x=x, y=y, blink=blink, quality1=qty1, quality2=qty2)
		# 		for (j in 1:length(band.names)) {
		# 			result[paste(ch1, ch2, "_coherence_mean_", band.names[j], "_power", sep="")] <- mean.coherence(cohr, bands[j,])
		# 		}
		# 	}
		# }
    
    	result
    
	} else {
		print(paste("File", file, "does not exist"))
	}
}

#' Download a summary file
#' @param filename the name of the raw data file
#' @preempt cors
#' @serializer contentType list(type="text/plain")
#' @post /v1/sumfile/
function(req, filename, sampling=128, window=2, sliding=0.75) {	
	channels <- c("AF3", "F7", "F3", "FC5", 
                "T7", "P7", "O1", "O2", 
                "P8", "T8", "FC6", "F4", 
                "F8", "AF4")
	print("obtaining user data from header:")
	print(paste0("xuser=",req$HTTP_X_USER))
	json <- fromJSON(req$HTTP_X_USER)
  	print(paste0("userName: ", json$userName))

	file <- paste("./raw-data/", json$userName, "/" , filename , ".txt", sep="")
  
	if ( file.exists(file) ) {
		data <- read.table(file, header=T)
		samples <- dim(data)[1]
		result <- list("Subject"=filename, "Version" = version, "Session" = "NA", "Sampling"=sampling,
						"Window"=window, "Sliding"=sliding, "Duration" = (samples / sampling), "Blinks" = "NA")
		
		if ("Blink" %in% names(data)) {
			blink <- data$Blink
			blink_onsets <- blink[2 : samples] - blink[1 : (samples - 1)]
			result["Meta_Blinks"] <- sum(blink_onsets[blink_onsets > 0])
		} else {
			blink <- rep(0, samples)
		}
		x <- data$GyroX[1 : samples]
		y <- data$GyroY[1 : samples]
		
		for (ch in channels) {
			#print(ch)
			ts <- data[[ch]]
			ts <- ts[1 : samples]
			qty <- data[[paste(ch, "Q", sep="_")]]
			qty <- qty[1 : samples]
			spectrum <- spectral.analysis(ts, sampling, length=window, sliding=0.75, hamming=T,
											x=x, y=y, blink=blink, quality=qty)
			
			for (j in 1:length(band.names)) {
				result[paste(ch, "_mean_", band.names[j], "_power", sep="")] <- mean.power(spectrum, bands[j,])
			}
			result[paste(ch, "IAF", sep="_")] <- iaf(spectrum)
			result[paste(ch, "IAF", "Power", sep="_")] <- iaf.power(spectrum)
			result[paste("Meta", ch, "Samples", sep="_")] <- spectrum$Samples
			result[paste("Meta", ch, "LongestQualitySegment", sep="_")] <- spectrum$LongestQualitySegment
			result[paste("Meta", ch, "SpectralQuality", sep="_")] <- spectral.quality(spectrum)
		}
		
		## Coherence analysis
		for (i in  1 : (length(channels) - 1)) {
			for (j in  (i + 1) : length(channels)) {
				ch1 <- channels[i]
				ch2 <- channels[j]
				
				ts1 <- data[[ch1]]
				ts2 <- data[[ch2]]
				
				ts1 <- ts1[1 : samples]
				ts2 <- ts2[1 : samples]
				
				qty1 <- data[[paste(ch1, "Q", sep="_")]]
				qty1 <- qty1[1 : samples]
				
				qty2 <- data[[paste(ch2, "Q", sep="_")]]
				qty2 <- qty2[1 : samples]
				#print(paste("Coherence", ch1, ch2))
				cohr <- coherence.analysis(ts1, ts2, sampling, length=window, sliding=0.75, hamming=T,
												x=x, y=y, blink=blink, quality1=qty1, quality2=qty2)
				for (j in 1:length(band.names)) {
					result[paste(ch1, ch2, "_coherence_mean_", band.names[j], "_power", sep="")] <- mean.coherence(cohr, bands[j,])
				}
			}
		}
    
    	tmp <- tempfile()
		write.table(as.data.frame(result), file=tmp,
		            quote=F, row.names=F, col.names=T, sep="\t")
		readBin(tmp, "raw", n=file.info(tmp)$size)
    
	} else {
		print(paste("File", file, "does not exist"))
	}
}


#' Create a spectogram for the given channel
#' @param ch The channel specified
#' @param subject The message to echo back.
#' @param session The session name
#' @preempt cors
#' @get /v1/spectrum/
#' @png
function(req, ch, subject, session, sampling=128, window=2, sliding=0.75) {	
  channels <- c("AF3", "F7", "F3", "FC5", 
                "T7", "P7", "O1", "O2", 
                "P8", "T8", "FC6", "F4", 
                "F8", "AF4")
  
	print("obtaining user data from header:")
	print(paste0("xuser=",req$HTTP_X_USER))
	json <- fromJSON(req$HTTP_X_USER)
	print(paste0("userName: ", json$userName))



	file <- paste("./raw-data/", json$userName, "/" ,subject, "_", session, ".txt", sep="")
  
  if ( file.exists(file) ) {
    data <- read.table(file, header=T)
    samples <- dim(data)[1]
    result <- list("Subject"=subject, "Version" = version, "Session"=session, "Sampling"=sampling,
                   "Window"=window, "Sliding"=sliding, "Duration" = (samples / sampling), "Blinks" = "NA")
    
    
    if ("Blink" %in% names(data)) {
      blink <- data$Blink
      blink_onsets <- blink[2 : samples] - blink[1 : (samples - 1)]
      result["Meta_Blinks"] <- sum(blink_onsets[blink_onsets > 0])
    } else {
      blink <- rep(0, samples)
    }
    x <- data$GyroX[1 : samples]
    y <- data$GyroY[1 : samples]
    
    textdata <- NULL     # Spectral text data
    c_textdata <- NULL   # Coherence text data
    
	ts <- data[[ch]]
	ts <- ts[1 : samples]
	qty <- data[[paste(ch, "Q", sep="_")]]
	qty <- qty[1 : samples]
	spectrum <- spectral.analysis(ts, sampling, length=window, sliding=0.75, hamming=T,
								x=x, y=y, blink=blink, quality=qty)
	
	if ( is.null(textdata) ) {
	textdata <- rbind(c("Subject", "Channel", paste(spectrum$Freq, "Hz", sep = "")), 
						c(subject, ch, spectrum$Spectrum))
	} else {
	textdata <- rbind(textdata, 
						c(subject, ch, spectrum$Spectrum))
	}
    
    plot.spectrum(spectrum, window, name=paste(subject, session, sep="/"), channel=ch)
    
  } else {
    print(paste("File", file, "does not exist"))
  }
}

#' Create a spectrum file
#' @param subject The subject name
#' @param session The session name
#' @preempt cors
#' @get /v1/specfile/
function(req, subject, session, sampling=128, window=2, sliding=0.75) {	
  channels <- c("AF3", "F7", "F3", "FC5", 
                "T7", "P7", "O1", "O2", 
                "P8", "T8", "FC6", "F4", 
                "F8", "AF4")
  
    print("obtaining user data from header:")
	print(paste0("xuser=",req$HTTP_X_USER))
	json <- fromJSON(req$HTTP_X_USER)
  	print(paste0("userName: ", json$userName))



	file <- paste("./raw-data/", json$userName, "/" ,subject, "_", session, ".txt", sep="")
  
  if ( file.exists(file) ) {
    data <- read.table(file, header=T)
    samples <- dim(data)[1]
    result <- list("Subject"=subject, "Version" = version, "Session"=session, "Sampling"=sampling,
                   "Window"=window, "Sliding"=sliding, "Duration" = (samples / sampling), "Blinks" = "NA")
    
    
    if ("Blink" %in% names(data)) {
      blink <- data$Blink
      blink_onsets <- blink[2 : samples] - blink[1 : (samples - 1)]
      result["Meta_Blinks"] <- sum(blink_onsets[blink_onsets > 0])
    } else {
      blink <- rep(0, samples)
    }
    x <- data$GyroX[1 : samples]
    y <- data$GyroY[1 : samples]
    
    textdata <- NULL     # Spectral text data
    c_textdata <- NULL   # Coherence text data
    
	for (ch in channels) {
		#print(ch)
		ts <- data[[ch]]
		ts <- ts[1 : samples]
		qty <- data[[paste(ch, "Q", sep="_")]]
		qty <- qty[1 : samples]
		spectrum <- spectral.analysis(ts, sampling, length=window, sliding=0.75, hamming=T,
									x=x, y=y, blink=blink, quality=qty)
		if ( is.null(textdata) ) {
		textdata <- rbind(c("Subject", "Channel", "IAF", "IAF_Power", "Sd", paste(spectrum$Freq, "Hz", sep = "")), 
							c(subject, ch, iaf(spectrum), iaf.power(spectrum),sd(spectrum$Spectrum), spectrum$Spectrum))
		} else {
		textdata <- rbind(textdata, 
							c(subject, ch, iaf(spectrum), iaf.power(spectrum),sd(spectrum$Spectrum), spectrum$Spectrum))
		
		}
	}

	textdata
    
  } else {
    print(paste("File", file, "does not exist"))
  }
}


#' Downloads a spectrum file
#' @param filename the name of the raw data file
#' @preempt cors
#' @serializer contentType list(type="text/plain")
#' @post /v1/specfile/
function(req, filename, sampling=128, window=2, sliding=0.75) {	
    channels <- c("AF3", "F7", "F3", "FC5", 
                "T7", "P7", "O1", "O2", 
                "P8", "T8", "FC6", "F4", 
                "F8", "AF4")
  
    print("obtaining user data from header:")
	print(paste0("xuser=",req$HTTP_X_USER))
	json <- fromJSON(req$HTTP_X_USER)
  	print(paste0("userName: ", json$userName))

	file <- paste("./raw-data/", json$userName, "/" , filename , ".txt", sep="")
  
  if ( file.exists(file) ) {
    data <- read.table(file, header=T)
    samples <- dim(data)[1]
    
    if ("Blink" %in% names(data)) {
      blink <- data$Blink
      blink_onsets <- blink[2 : samples] - blink[1 : (samples - 1)]
    } else {
      blink <- rep(0, samples)
    }
    x <- data$GyroX[1 : samples]
    y <- data$GyroY[1 : samples]
    
    textdata <- NULL     # Spectral text data
    
	for (ch in channels) {
		#print(ch)
		ts <- data[[ch]]
		ts <- ts[1 : samples]
		qty <- data[[paste(ch, "Q", sep="_")]]
		qty <- qty[1 : samples]
		spectrum <- spectral.analysis(ts, sampling, length=window, sliding=0.75, hamming=T,
									x=x, y=y, blink=blink, quality=qty)
		if ( is.null(textdata) ) {
		textdata <- rbind(c("Subject", "Channel", "IAF", "IAF_Power", "Sd", paste(spectrum$Freq, "Hz", sep = "")), 
							c(filename, ch, iaf(spectrum), iaf.power(spectrum),sd(spectrum$Spectrum), spectrum$Spectrum))
		} else {
		textdata <- rbind(textdata, 
							c(filename, ch, iaf(spectrum), iaf.power(spectrum),sd(spectrum$Spectrum), spectrum$Spectrum))
		
		}
	}

	tmp <- tempfile()
	write.table(textdata, file=tmp,
				quote=F, row.names=F, col.names=F, sep="\t")
	readBin(tmp, "raw", n=file.info(tmp)$size)
    
  } else {
    print(paste("File", file, "does not exist"))
  }
}

#' Create a coherence file
#' @param subject The subject name
#' @param session The session name
#' @preempt cors
#' @get /v1/cohrfile/
function(req, subject, session, sampling=128, window=2, sliding=0.75) {	
  channels <- c("AF3", "F7", "F3", "FC5", 
                "T7", "P7", "O1", "O2", 
                "P8", "T8", "FC6", "F4", 
                "F8", "AF4")
  
    print("obtaining user data from header:")
	print(paste0("xuser=",req$HTTP_X_USER))
	json <- fromJSON(req$HTTP_X_USER)
  	print(paste0("userName: ", json$userName))

	file <- paste("./raw-data/", json$userName, "/" ,subject, "_", session, ".txt", sep="")
  
  if ( file.exists(file) ) {
    data <- read.table(file, header=T)
    samples <- dim(data)[1]
    result <- list("Subject"=subject, "Version" = version, "Session"=session, "Sampling"=sampling,
                   "Window"=window, "Sliding"=sliding, "Duration" = (samples / sampling), "Blinks" = "NA")
    
    
    if ("Blink" %in% names(data)) {
      blink <- data$Blink
      blink_onsets <- blink[2 : samples] - blink[1 : (samples - 1)]
      result["Meta_Blinks"] <- sum(blink_onsets[blink_onsets > 0])
    } else {
      blink <- rep(0, samples)
    }
    x <- data$GyroX[1 : samples]
    y <- data$GyroY[1 : samples]
    
    c_textdata <- NULL   # Coherence text data
    
	## Coherence analysis
	for (i in  1 : (length(channels) - 1)) {
		for (j in  (i + 1) : length(channels)) {
			ch1 <- channels[i]
			ch2 <- channels[j]
			
			ts1 <- data[[ch1]]
			ts2 <- data[[ch2]]
			
			ts1 <- ts1[1 : samples]
			ts2 <- ts2[1 : samples]
			
			qty1 <- data[[paste(ch1, "Q", sep="_")]]
			qty1 <- qty1[1 : samples]
			
			qty2 <- data[[paste(ch2, "Q", sep="_")]]
			qty2 <- qty2[1 : samples]
			# print(paste("Coherence", ch1, ch2))
			cohr <- coherence.analysis(ts1, ts2, sampling, length=window, sliding=0.75, hamming=T,
											x=x, y=y, blink=blink, quality1=qty1, quality2=qty2)
			for (j in 1:length(band.names)) {
				result[paste(ch1, ch2, "_coherence_mean_", band.names[j], "_power", sep="")] <- mean.coherence(cohr, bands[j,])
			}
			
			## Update coherence table
			
			if ( is.null(c_textdata) ) {
				c_textdata <- rbind(c("Subject", "Channel1", "Channel2", paste(cohr$Freq, "Hz", sep = "")), 
								c(subject, ch1, ch2, cohr$Coherence))
			} else {
				c_textdata <- rbind(c_textdata, 
								c(subject, ch1, ch2, cohr$Coherence))
			}
		}
	}
	
	c_textdata
    
  } else {
    print(paste("File", file, "does not exist"))
  }
}

#' Downloads a coherence file
#' @param filename the name of the raw data file
#' @preempt cors
#' @serializer contentType list(type="text/plain")
#' @post /v1/cohrfile/
function(req, filename, sampling=128, window=2, sliding=0.75) {	
  channels <- c("AF3", "F7", "F3", "FC5", 
                "T7", "P7", "O1", "O2", 
                "P8", "T8", "FC6", "F4", 
                "F8", "AF4")
  
    print("obtaining user data from header:")
	print(paste0("xuser=",req$HTTP_X_USER))
	json <- fromJSON(req$HTTP_X_USER)
  	print(paste0("userName: ", json$userName))

	file <- paste("./raw-data/", json$userName, "/" , filename , ".txt", sep="")
  
  if ( file.exists(file) ) {
    data <- read.table(file, header=T)
    samples <- dim(data)[1]
    
    if ("Blink" %in% names(data)) {
      blink <- data$Blink
      blink_onsets <- blink[2 : samples] - blink[1 : (samples - 1)]
    } else {
      blink <- rep(0, samples)
    }
    x <- data$GyroX[1 : samples]
    y <- data$GyroY[1 : samples]
    
    c_textdata <- NULL   # Coherence text data
    
	## Coherence analysis
	for (i in  1 : (length(channels) - 1)) {
		for (j in  (i + 1) : length(channels)) {
			ch1 <- channels[i]
			ch2 <- channels[j]
			
			ts1 <- data[[ch1]]
			ts2 <- data[[ch2]]
			
			ts1 <- ts1[1 : samples]
			ts2 <- ts2[1 : samples]
			
			qty1 <- data[[paste(ch1, "Q", sep="_")]]
			qty1 <- qty1[1 : samples]
			
			qty2 <- data[[paste(ch2, "Q", sep="_")]]
			qty2 <- qty2[1 : samples]
			# print(paste("Coherence", ch1, ch2))
			cohr <- coherence.analysis(ts1, ts2, sampling, length=window, sliding=0.75, hamming=T,
											x=x, y=y, blink=blink, quality1=qty1, quality2=qty2)

			## Update coherence table
			
			if ( is.null(c_textdata) ) {
				c_textdata <- rbind(c("Subject", "Channel1", "Channel2", paste(cohr$Freq, "Hz", sep = "")), 
								c(filename, ch1, ch2, cohr$Coherence))
			} else {
				c_textdata <- rbind(c_textdata, 
								c(filename, ch1, ch2, cohr$Coherence))
			}
		}
	}
	
	tmp <- tempfile()
	write.table(c_textdata, file=tmp,
				quote=F, row.names=F, col.names=F, sep="\t")
	readBin(tmp, "raw", n=file.info(tmp)$size)
    
  } else {
    print(paste("File", file, "does not exist"))
  }
}


#' Return a cleaning template of the summary data
#' @param filename filename
#' @preempt cors
#' @serializer contentType list(type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
#' @get /v1/clean/
function(req, filename, sampling=128, window=2, sliding=0.75) {	
  channels <- c("AF3", "F7", "F3", "FC5", 
                "T7", "P7", "O1", "O2", 
                "P8", "T8", "FC6", "F4", 
                "F8", "AF4")
  
    print("obtaining user data from header:")
	print(paste0("xuser=",req$HTTP_X_USER))
	json <- fromJSON(req$HTTP_X_USER)
  	print(paste0("userName: ", json$userName))

	file <- paste("./raw-data/", json$userName, "/" , filename , ".txt", sep="")
  
  if ( file.exists(file) ) {
    data <- read.table(file, header=T)
    samples <- dim(data)[1]
    result <- list("Subject"=filename, "Version" = version, "Session"="NA", "Sampling"=sampling,
                   "Window"=window, "Sliding"=sliding, "Duration" = (samples / sampling), "Blinks" = "NA")
    
    
    if ("Blink" %in% names(data)) {
      blink <- data$Blink
      blink_onsets <- blink[2 : samples] - blink[1 : (samples - 1)]
      result["Meta_Blinks"] <- sum(blink_onsets[blink_onsets > 0])
    } else {
      blink <- rep(0, samples)
    }
    x <- data$GyroX[1 : samples]
    y <- data$GyroY[1 : samples]
    
    
    for (ch in channels) {
      #print(ch)
      ts <- data[[ch]]
      ts <- ts[1 : samples]
      qty <- data[[paste(ch, "Q", sep="_")]]
      qty <- qty[1 : samples]
      spectrum <- spectral.analysis(ts, sampling, length=window, sliding=0.75, hamming=T,
                                    x=x, y=y, blink=blink, quality=qty)
      
      for (j in 1:length(band.names)) {
        result[paste(ch, "_mean_", band.names[j], "_power", sep="")] <- mean.power(spectrum, bands[j,])
      }
      result[paste(ch, "IAF", sep="_")] <- iaf(spectrum)
      result[paste(ch, "IAF", "Power", sep="_")] <- iaf.power(spectrum)
      result[paste("Meta", ch, "Samples", sep="_")] <- spectrum$Samples
      result[paste("Meta", ch, "LongestQualitySegment", sep="_")] <- spectrum$LongestQualitySegment
      result[paste("Meta", ch, "SpectralQuality", sep="_")] <- spectral.quality(spectrum)
    }
    
    ## Coherence analysis
    for (i in  1 : (length(channels) - 1)) {
      for (j in  (i + 1) : length(channels)) {
        ch1 <- channels[i]
        ch2 <- channels[j]
        
        ts1 <- data[[ch1]]
        ts2 <- data[[ch2]]
        
        ts1 <- ts1[1 : samples]
        ts2 <- ts2[1 : samples]
        
        qty1 <- data[[paste(ch1, "Q", sep="_")]]
        qty1 <- qty1[1 : samples]
        
        qty2 <- data[[paste(ch2, "Q", sep="_")]]
        qty2 <- qty2[1 : samples]
        #print(paste("Coherence", ch1, ch2))
        cohr <- coherence.analysis(ts1, ts2, sampling, length=window, sliding=0.75, hamming=T,
                                   x=x, y=y, blink=blink, quality1=qty1, quality2=qty2)
        for (j in 1:length(band.names)) {
          result[paste(ch1, ch2, "_coherence_mean_", band.names[j], "_power", sep="")] <- mean.coherence(cohr, bands[j,])
        }
      }
    }

	Sys.getenv("R_ZIPCMD")
	Sys.setenv(R_ZIPCMD = "zip")
    
    raw <- getSheetNames("eeg_template.xlsx")[[1]]
	wb <- loadWorkbook("eeg_template.xlsx")
    writeData(wb, raw, as.data.frame(t(as.matrix(result))), rowNames = TRUE)
    saveWorkbook(wb, "eeg_template.xlsx", overwrite = TRUE)
    # code <- base64encode("eeg_template.xlsx")
	# output <- list("format" = "xlsx", "docTitle" = "eeg_template", "document" = code)
	# output

	filename <- "eeg_template.xlsx"
	readBin(filename, "raw", n=file.info(filename)$size)
	
  } else {
    print(paste("File", file, "does not exist"))
  }
}
